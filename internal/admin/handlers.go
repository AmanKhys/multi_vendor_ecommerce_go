package admin

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/validators"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	log "github.com/sirupsen/logrus"
)

type Admin struct{ DB *db.Queries }

func (a *Admin) AdminAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	var resp struct {
		Data    []db.GetAllUsersRow `json:"data"`
		Message string              `json:"message"`
	}
	data, err := a.DB.GetAllUsers(context.TODO())
	if err == sql.ErrNoRows {
		message := "no current users available to display"
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(message))
	} else if err != nil {
		log.Warn(err)
		http.Error(w, fmt.Errorf("failed : %w", err).Error(), http.StatusBadRequest)
		return
	}

	resp.Data = data
	resp.Message = "successfully fetched all users"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (a *Admin) AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	var resp struct {
		Data    []db.GetAllUsersByRoleUserRow `json:"data"`
		Message string                        `json:"message"`
	}
	data, err := a.DB.GetAllUsersByRoleUser(context.TODO(), "user")
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "text/plain")
		message := "no current users available"
		w.Write([]byte(message))
	} else if err != nil {
		log.Warn(err)
		http.Error(w, fmt.Errorf("failed : %w", err).Error(), http.StatusBadRequest)
		return
	}

	resp.Data = data
	resp.Message = "successfully fetched all users"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (a *Admin) AdminSellersHandler(w http.ResponseWriter, r *http.Request) {
	var resp struct {
		Data    []db.GetAllUsersByRoleSellerRow `json:"data"`
		Message string                          `json:"message"`
	}
	data, err := a.DB.GetAllUsersByRoleSeller(context.TODO(), "seller")
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "text/plain")
		message := "no sellers available"
		w.Write([]byte(message))
	} else if err != nil {
		log.Warn(err)
		http.Error(w, fmt.Errorf("failed : %w", err).Error(), http.StatusBadRequest)
		return
	}

	resp.Data = data
	resp.Message = "successfully fetched all sellers"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}
func (a *Admin) VerifySellerHandler(w http.ResponseWriter, r *http.Request) {
	// check if the request is in correct format
	var req struct {
		Email string `json:"email"`
	}
	req.Email = r.URL.Query().Get("email")
	if !validators.ValidateEmail(req.Email) {
		http.Error(w, "invalid email format:", http.StatusBadRequest)
		return
	}

	// check if the seller exists and is not verified
	user, err := a.DB.GetUserByEmail(context.TODO(), req.Email)
	if err == sql.ErrNoRows {
		http.Error(w, "seller does not exist.", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error fetching user by email")
		http.Error(w, "internal server error while fetching user", http.StatusInternalServerError)
		return
	} else if user.Role != utils.SellerRole {
		http.Error(w, "the given details is not that of a seller", http.StatusBadRequest)
		return
	} else if user.UserVerified {
		http.Error(w, "seller already verified", http.StatusBadRequest)
		return
	} else if !user.EmailVerified {
		http.Error(w, "seller email not yet verified.", http.StatusUnauthorized)
		return
	}

	// verify seller
	seller, err := a.DB.VerifySellerByID(context.TODO(), user.ID)
	if err != nil {
		log.Warn("verify seller by id failed for a valid seller.")
		http.Error(w, "internal server error while verifying seller", http.StatusInternalServerError)
		return
	}
	// make errors and messages slice for the response
	var Err []string
	var Messages []string
	// add wallet for the seller
	wallet, err := a.DB.AddWalletByUserID(context.TODO(), seller.ID)
	if err != nil {
		log.Warn("error adding wallet for seller after verifying account:", err.Error())
		Err = append(Err, "error adding wallet for seller after verifying account")
	} else {
		Messages = append(Messages, "successfully added wallet for seller")
		Messages = append(Messages, "walletID:", wallet.ID.String(), fmt.Sprintf("savings: %v", wallet.Savings))
	}
	var resp struct {
		Data     db.VerifySellerByIDRow `json:"data"`
		Messages []string               `json:"message"`
		Err      []string               `json:"errors"`
	}
	resp.Data = seller
	resp.Messages = append(Messages, "successfully verified seller")
	resp.Err = Err
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (a *Admin) AdminProductsHandler(w http.ResponseWriter, r *http.Request) {
	var resp struct {
		Data    []db.Product `json:"data"`
		Message string       `json:"message"`
	}
	data, err := a.DB.GetAllProductsForAdmin(context.TODO())
	if err != nil {
		log.Warn(err)
		http.Error(w, fmt.Errorf("failed : %w", err).Error(), http.StatusBadRequest)
		return
	}

	resp.Data = data
	resp.Message = "successfully fetched all products"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}

func (a *Admin) AdminCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	var resp struct {
		Data    []db.Category `json:"data"`
		Message string        `json:"message"`
	}
	data, err := a.DB.GetAllCategoriesForAdmin(context.TODO())
	if err != nil {
		log.Warn(err)
		http.Error(w, fmt.Errorf("failed : %w", err).Error(), http.StatusBadRequest)
		return
	}

	resp.Data = data
	resp.Message = "successfully fetched all categories"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (a *Admin) BlockUserHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserIDStr string `json:"user_id"`
	}
	req.UserIDStr = r.URL.Query().Get("user_id")
	userID, err := uuid.Parse(req.UserIDStr)
	if err != nil {
		http.Error(w, "userID not in valid format", http.StatusBadRequest)
		return
	}

	user, err := a.DB.GetUserById(context.TODO(), userID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid userID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error taking user from db: ", err)
		http.Error(w, "error fetching user", http.StatusInternalServerError)
		return
	} else if user.IsBlocked {
		http.Error(w, "user already blocked", http.StatusBadRequest)
		return
	} else if user.Role == utils.AdminRole {
		http.Error(w, "trying to block admin: invalid request", http.StatusBadRequest)
		return
	}

	blockedUser, err := a.DB.BlockUserByID(context.TODO(), userID)
	if err != nil {
		log.Warn(err)
		http.Error(w, "error blocking user", http.StatusInternalServerError)
		return
	}
	log.Infof("blocked user: %s", blockedUser.Email)
	message := fmt.Sprintf("succesfully blocked user: %s", blockedUser.ID.String())
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(message))
}

func (a *Admin) UnblockUserHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserIDStr string `json:"user_id"`
	}
	req.UserIDStr = r.URL.Query().Get("user_id")
	userID, err := uuid.Parse(req.UserIDStr)
	if err != nil {
		http.Error(w, "invalid userID format", http.StatusBadRequest)
		return
	}

	user, err := a.DB.UnblockUserByID(context.TODO(), userID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid user data", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn(err)
		http.Error(w, "error unblocking user", http.StatusInternalServerError)
		return
	}
	log.Infof("unblocked user: %s", user.ID.String())
	message := fmt.Sprintf("succesfully unblocked user: %s", user.ID.String())
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(message))
}

func (a *Admin) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductID uuid.UUID `json:"id"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}
	product, err := a.DB.DeleteProductByID(context.TODO(), req.ProductID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid productID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn(err)
		http.Error(w, "error deleting product", http.StatusInternalServerError)
		return
	}
	log.Infof("deleted product: %s", product.ID.String())
	w.Header().Set("Content-Type", "application/json")
	message := fmt.Sprintf("product: %s deleted", product.Name)
	w.Write([]byte(message))
}

func (a *Admin) AddCategoryHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}
	req.Name = strings.ToLower(req.Name)
	category, err := a.DB.AddCateogry(context.TODO(), req.Name)
	if err != nil {
		log.Warn(err)
		http.Error(w, fmt.Errorf("failed to add cateogry: %w", err).Error(), http.StatusBadRequest)
		return
	}
	log.Infof("added category: %s", category.Name)
	w.Header().Set("Content-Type", "text/plain")
	message := fmt.Sprintf("category: %s added", category.Name)
	w.Write([]byte(message))
}

func (a *Admin) EditCategoryHandler(w http.ResponseWriter, r *http.Request) {
	var req db.EditCategoryNameByNameParams
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}
	req.Name = strings.ToLower(req.Name)
	req.NewName = strings.ToLower(req.NewName)
	category, err := a.DB.EditCategoryNameByName(context.TODO(), req)
	if err != nil {
		log.Warn(err)
		http.Error(w, fmt.Errorf("failed to rename cateogry: %w", err).Error(), http.StatusBadRequest)
		return
	}
	log.Infof("renamed category: %s", category.Name)
	w.Header().Set("Content-Type", "text/plain")
	message := fmt.Sprintf("category: %s renamed to %s", req.Name, category.Name)
	w.Write([]byte(message))
}

func (a *Admin) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CategoryName string `json:"name"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}
	req.CategoryName = strings.ToLower(req.CategoryName)
	category, err := a.DB.DeleteCategoryByName(context.TODO(), req.CategoryName)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid category name", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, fmt.Errorf("failed to delete category: %w", err).Error(), http.StatusBadRequest)
		return
	}
	log.Infof("deleted category: %s", category.Name)
	w.Header().Set("Content-Type", "text/plain")
	message := fmt.Sprintf("category: %s deleted", category.Name)
	w.Write([]byte(message))
}

func (a *Admin) GetOrderItemsHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	// fetch order_items
	orderItems, err := a.DB.GetAllOrderItemsForAdmin(context.TODO())
	if err != nil {
		log.Warn("error fetching orders for admin in GetOrderItemsHandler:", err.Error())
		http.Error(w, "internal server error fetching orders for admin", http.StatusInternalServerError)
		return
	}

	// make resp orderItem struct
	type respOrderItem struct {
		ID          uuid.UUID `json:"order_item_id"`
		OrderID     uuid.UUID `json:"order_id"`
		Status      string    `json:"status"`
		ProductID   uuid.UUID `json:"product_id"`
		Price       float64   `json:"price"`
		Quantity    int       `json:"quantity"`
		TotalAmount float64   `json:"total_amount"`
		CreatedAt   time.Time `json:"created_at"`
	}

	// respOrderItems slice
	var respOrderItems []respOrderItem
	for _, v := range orderItems {
		var temp respOrderItem
		temp.ID = v.ID
		temp.OrderID = v.OrderID
		temp.Status = v.Status
		temp.ProductID = v.ProductID
		temp.Price = v.Price
		temp.Quantity = int(v.Quantity)
		temp.TotalAmount = v.TotalAmount
		temp.CreatedAt = v.CreatedAt

		respOrderItems = append(respOrderItems, temp)
	}

	// give response
	var resp struct {
		Data    []respOrderItem `json:"data"`
		Message string          `json:"message"`
	}
	resp.Data = respOrderItems
	resp.Message = "successfully fetched all orderItems"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (a *Admin) DeliverOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	// take request value from params
	var req struct {
		OrderItemIDStr string
	}
	req.OrderItemIDStr = r.URL.Query().Get("order_item_id")
	orderID, err := uuid.Parse(req.OrderItemIDStr)
	if err != nil {
		http.Error(w, "not a valid orderItemID", http.StatusBadRequest)
		return
	}

	// fetch orderItemByID
	orderItem, err := a.DB.GetOrderItemByID(context.TODO(), orderID)
	if err == sql.ErrNoRows {
		http.Error(w, "not a valid orderItemID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error fetching orderItemByID in admin to change orderStatus:", err.Error())
		http.Error(w, "internal server error fetching orderItem", http.StatusInternalServerError)
		return
	}

	// checking order item status to change the status to delivered
	if orderItem.Status == utils.StatusOrderCancelled ||
		orderItem.Status == utils.StatusOrderPending ||
		orderItem.Status == utils.StatusOrderProcessing ||
		orderItem.Status == utils.StatusOrderDelivered {
		msg := fmt.Sprintf("order %s. Cannot change to status to delivered. can only deliver orderItem that is shipped", orderItem.Status)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}
	// no need to check the orderItem status is shipped since it is the
	// default case if all the cases above failed

	var arg db.ChangeOrderItemStatusByIDParams
	arg.ID = orderItem.ID
	arg.Status = utils.StatusOrderDelivered
	updatedOrderItem, err := a.DB.ChangeOrderItemStatusByID(context.TODO(), arg)
	if err != nil {
		log.Warn("error changing order status to delivered in DeliverOrderItemHandler in admin:", err.Error())
		http.Error(w, "internal error changing order status to delivered", http.StatusInternalServerError)
		return
	}
	// if updatedOrderItem.Status == utils.Status
	// var editVendorPayArg db.EditVendorPaymentStatusByOrderItemIDParams
	// editVendorPayArg.OrderItemID = updatedOrderItem.ID
	// editVendorPayArg.Status = utils.StatusVendorPaymentReceived
	// a.DB.EditVendorPaymentStatusByOrderItemID(context.TODO(), editVendorPayArg)
	type respOrderItem struct {
		OrderItemID uuid.UUID `json:"order_item_id"`
		Status      string    `json:"status"`
		ProductID   uuid.UUID `json:"product_id"`
		Price       float64   `json:"price"`
		Qauntity    int       `json:"quantity"`
		TotalAmount float64   `json:"total_amount"`
	}
	var respUpdatedOrderItem respOrderItem
	respUpdatedOrderItem.OrderItemID = updatedOrderItem.ID
	respUpdatedOrderItem.ProductID = updatedOrderItem.ProductID
	respUpdatedOrderItem.Price = updatedOrderItem.Price
	respUpdatedOrderItem.Qauntity = int(updatedOrderItem.Quantity)
	respUpdatedOrderItem.TotalAmount = updatedOrderItem.TotalAmount
	respUpdatedOrderItem.Status = updatedOrderItem.Status
	var resp struct {
		Data    respOrderItem `json:"data"`
		Message string
	}
	resp.Data = respUpdatedOrderItem
	resp.Message = "successfully updated the orderItem to status delivered"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// coupon handlers
func (a *Admin) AdminCouponsHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	coupons, err := a.DB.GetAllCoupons(context.TODO())
	if err != nil {
		log.Error("error fetching coupons in AdminCouponsHandler:", err.Error())
		http.Error(w, "internal error fetching coupons for admin", http.StatusInternalServerError)
		return
	}
	type respCoupon struct {
		ID             uuid.UUID `json:"id"`
		Name           string    `json:"name"`
		TriggerPrice   float64   `json:"trigger_price"`
		DiscountAmount float64   `json:"discount_amount"`
	}

	var respCoupons []respCoupon
	for _, v := range coupons {
		var temp respCoupon
		temp.ID = v.ID
		temp.Name = v.Name
		temp.TriggerPrice = v.TriggerPrice
		temp.DiscountAmount = v.DiscountAmount
		respCoupons = append(respCoupons, temp)
	}

	var resp struct {
		Data    []respCoupon `json:"data"`
		Message string       `json:"messsage"`
	}
	resp.Data = respCoupons
	resp.Message = "successfully fetched all coupons"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
func (a *Admin) AddCouponHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	var req struct {
		Name           string  `json:"name"`
		TriggerPrice   float64 `json:"trigger_price"`
		DiscountAmount float64 `json:"discount_amount"`
	}

	var errors []string
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "wrong request body format", http.StatusBadRequest)
		return
	}
	if !validators.ValidateCouponName(req.Name) {
		errors = append(errors, "invalid name")
	}
	if !validators.ValidateCouponPrice(req.TriggerPrice) {
		errors = append(errors, "invalid trigger price")
	}
	if !validators.ValidateCouponPrice(req.DiscountAmount) {
		errors = append(errors, "invalid discount amount")
	}
	if req.TriggerPrice <= req.DiscountAmount {
		errors = append(errors, "error: trigger price less than or equal to discount amount")
	}

	if len(errors) > 0 {
		http.Error(w, strings.Join(errors, "\n"), http.StatusBadRequest)
		return
	}

	var addCouponArg db.AddCouponParams
	addCouponArg.Name = req.Name
	addCouponArg.TriggerPrice = req.TriggerPrice
	addCouponArg.DiscountAmount = req.DiscountAmount
	addedCoupon, err := a.DB.AddCoupon(context.TODO(), addCouponArg)
	if err != nil {
		log.Error("error adding coupon after successful validation:", err.Error())
		http.Error(w, "internal error adding coupon", http.StatusInternalServerError)
		return
	}

	type respCoupon struct {
		Name           string  `json:"name"`
		TriggerPrice   float64 `json:"trigger_price"`
		DiscountAmount float64 `json:"discount_amount"`
	}
	var respCouponData respCoupon
	respCouponData.Name = addedCoupon.Name
	respCouponData.TriggerPrice = addedCoupon.TriggerPrice
	respCouponData.DiscountAmount = addedCoupon.DiscountAmount
	var resp struct {
		Data    respCoupon `json:"data"`
		Message string     `json:"message"`
	}
	resp.Data = respCouponData
	resp.Message = "successfully added coupon"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (a *Admin) EditCouponHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	// get and validate request body values
	var req struct {
		OldName        string  `json:"old_name"`
		NewName        string  `json:"new_name"`
		TriggerPrice   float64 `json:"trigger_price"`
		DiscountAmount float64 `json:"discount_amount"`
	}
	var errors []string
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "wrong json request format", http.StatusBadRequest)
		return
	}
	if !validators.ValidateCouponName(req.OldName) {
		errors = append(errors, "invalid coupon name to edit")
	}
	if !validators.ValidateCouponName(req.NewName) {
		errors = append(errors, "invalid new coupon name")
	}
	if !validators.ValidateCouponPrice(req.TriggerPrice) {
		errors = append(errors, "invalid trigger price")
	}
	if !validators.ValidateCouponPrice(req.DiscountAmount) {
		errors = append(errors, "invalid discount price")
	}
	if req.DiscountAmount >= req.TriggerPrice {
		errors = append(errors, "not allowed: discount price more than or equal to trigger price")
	}
	if len(errors) > 0 {
		http.Error(w, strings.Join(errors, "\n"), http.StatusBadRequest)
		return
	}

	// // fetch if the coupon exists
	// coupon, err := a.DB.GetCouponByName(context.TODO(), req.OldName)
	// if err == sql.ErrNoRows {
	// 	http.Error(w, "coupon does not exist", http.StatusBadRequest)
	// 	return
	// } else if err != nil {
	// 	log.Error("error fetching coupon to edit in EditCouponHandler in Admin:", err.Error())
	// 	http.Error(w, "internal error fetching coupon to edit", http.StatusInternalServerError)
	// 	return
	// }

	var editCouponArg db.EditCouponByNameParams
	editCouponArg.OldName = req.OldName
	editCouponArg.NewName = req.NewName
	editCouponArg.TriggerPrice = req.TriggerPrice
	editCouponArg.DiscountAmount = req.DiscountAmount
	editedCoupon, err := a.DB.EditCouponByName(context.TODO(), editCouponArg)
	if err == sql.ErrNoRows {
		http.Error(w, "coupon does not exist to edit", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Error("error updating editCouponByName:", err.Error())
		http.Error(w, "internal error editing coupon", http.StatusInternalServerError)
		return
	}

	var data struct {
		CouponID       uuid.UUID `json:"coupon_id"`
		NewName        string    `json:"new_name"`
		OldName        string    `json:"old_name"`
		TriggerPrice   float64   `json:"trigger_price"`
		DiscountAmount float64   `json:"discount_amount"`
		Message        string    `json:"message"`
	}
	data.CouponID = editedCoupon.ID
	data.NewName = editedCoupon.Name
	data.OldName = editCouponArg.OldName
	data.TriggerPrice = editedCoupon.TriggerPrice
	data.DiscountAmount = editedCoupon.DiscountAmount
	data.Message = "successfully updated coupon"

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)

}

func (a *Admin) DeleteCouponHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	couponName := r.URL.Query().Get("coupon_name")
	coupon, err := a.DB.DeleteCouponByName(context.TODO(), couponName)
	if err != nil {
		log.Error("error soft deleting coupon:", err.Error())
		http.Error(w, "internal error soft deleting coupon", http.StatusInternalServerError)
		return
	} else if coupon.IsDeleted {
		msg := fmt.Sprintf("coupon %s already deleted.", coupon.Name)
		http.Error(w, msg, http.StatusForbidden)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	var resp struct {
		CouponName string `josn:"coupon_name"`
		Message    string `json:"message"`
	}
	resp.CouponName = coupon.Name
	resp.Message = "coupon has been successfully deleted."
	json.NewEncoder(w).Encode(resp)

}

func (a *Admin) SalesReportHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid start_date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid end_date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	dateArg := db.GetVendorPaymentsByDateRangeParams{
		StartDate: startDate,
		EndDate:   endDate,
	}
	vendorPayments, err := a.DB.GetVendorPaymentsByDateRange(context.TODO(), dateArg)
	if err != nil {
		log.Error("error fetching vendorPayments:", err.Error())
		http.Error(w, "internal error fetching data", http.StatusInternalServerError)
		return
	}

	var totalProfit float64
	var totalSales int
	var totalCancelledOrders int
	orderItemMap := make(map[uuid.UUID]map[string]float64)

	for _, vp := range vendorPayments {
		if vp.Status == utils.StatusVendorPaymentCancelled {
			totalCancelledOrders++
			continue
		}
		totalSales++
		totalProfit += vp.PlatformFee

		if _, exists := orderItemMap[vp.OrderItemID]; !exists {
			orderItemMap[vp.OrderItemID] = map[string]float64{
				"sales":        0,
				"platform_fee": 0,
				"profit":       0,
			}
		}
		orderItemMap[vp.OrderItemID]["sales"] += vp.TotalAmount
		orderItemMap[vp.OrderItemID]["platform_fee"] += vp.PlatformFee
		orderItemMap[vp.OrderItemID]["profit"] += vp.CreditAmount
	}

	orders, err := a.DB.GetAllOrders(context.TODO())
	if err != nil {
		log.Error("error fetching orders:", err.Error())
		http.Error(w, "internal error fetching orders", http.StatusInternalServerError)
		return
	}

	var totalLossAmount float64
	for _, o := range orders {
		if o.CreatedAt.Before(startDate) || o.CreatedAt.After(endDate) {
			continue
		}
		payment, err := a.DB.GetPaymentByOrderID(context.TODO(), o.ID)
		if err != nil || payment.Status != utils.StatusPaymentSuccessful {
			continue
		}
		totalLossAmount += o.DiscountAmount
	}

	// Generate PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Admin Sales Report")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(95, 10, fmt.Sprintf("Start Date: %s", startDate.Format("2006-01-02")))
	pdf.Cell(95, 10, fmt.Sprintf("End Date: %s", endDate.Format("2006-01-02")))
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "Summary")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(95, 8, fmt.Sprintf("Total Sales: %d", totalSales))
	pdf.Cell(95, 8, fmt.Sprintf("Total Cancelled Orders: %d", totalCancelledOrders))
	pdf.Ln(8)
	pdf.Cell(95, 8, fmt.Sprintf("Total Platform Profit: $%.2f", totalProfit))
	pdf.Cell(95, 8, fmt.Sprintf("Total Loss by Discounts: $%.2f", totalLossAmount))
	pdf.Ln(8)
	pdf.Cell(190, 8, fmt.Sprintf("Net Profit: $%.2f", totalProfit-totalLossAmount))
	pdf.Ln(10)

	// Table Header
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(75, 8, "OrderItem ID", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 8, "Order Amount ($)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 8, "Platform Fee ($)", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(240, 240, 240)
	fill := false

	for id, data := range orderItemMap {
		pdf.CellFormat(75, 8, id.String(), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(50, 8, fmt.Sprintf("%.2f", data["sales"]), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(50, 8, fmt.Sprintf("%.2f", data["platform_fee"]), "1", 1, "C", fill, 0, "")
		fill = !fill
	}

	pdf.AddPage()
	// Table Header
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(100, 8, "Order ID", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 8, "Discount Amount($)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Coupon Name", "1", 1, "C", true, 0, "")

	// cells design
	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(240, 240, 240)
	fill = false
	for _, o := range orders {
		pdf.CellFormat(100, 8, o.ID.String(), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(50, 8, strconv.FormatFloat(o.DiscountAmount, 'f', -1, 64), "1", 0, "C", fill, 0, "")
		if !o.CouponID.Valid {
			pdf.CellFormat(40, 8, "no coupon", "1", 1, "C", fill, 0, "")
			fill = !fill
			continue
		}
		coupon, err := a.DB.GetCouponByID(context.TODO(), o.CouponID.UUID)
		if err != nil {
			log.Error("error fetching coupon by couponID in SalesReportHandler:", err.Error())
			pdf.CellFormat(40, 8, "no coupon", "1", 1, "C", fill, 0, "")
			fill = !fill
			continue
		}
		pdf.CellFormat(40, 8, coupon.Name, "1", 1, "C", fill, 0, "")
		fill = !fill
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		http.Error(w, "Error generating PDF", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=admin_sales_report.pdf")
	w.Write(buf.Bytes())
}
