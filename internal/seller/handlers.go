package seller

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

type Seller struct {
	DB *db.Queries
}

// edit profile handler
func (s *Seller) EditProfileHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	// get request
	var req struct {
		Name  string
		About string
		Phone string
	}
	req.Name = r.URL.Query().Get("name")
	req.About = r.URL.Query().Get("about")
	req.Phone = r.URL.Query().Get("phone")

	// make errors slice for response
	var Err []string
	if !validators.ValidateName(req.Name) && req.Name != "" {
		Err = append(Err, "invalid name format")
	}
	if !validators.ValidatePhone(req.Phone) && req.Phone != "" {
		Err = append(Err, "invalid phone number format")
	}
	if len(Err) > 0 {
		http.Error(w, strings.Join(Err, "\n"), http.StatusBadRequest)
		return
	}

	// update profile
	// check and make the update argument for db function
	var arg db.EditSellerByIDParams
	arg.ID = user.ID
	if req.Name == "" {
		arg.Name = user.Name
	} else {
		arg.Name = req.Name
	}
	if req.Phone == "" {
		arg.Phone = user.Phone
	} else {
		phoneInt, err := strconv.Atoi(req.Phone)
		if err != nil {
			arg.Phone = user.Phone
		} else {
			arg.Phone = sql.NullInt64{
				Int64: int64(phoneInt),
				Valid: true,
			}
		}
	}
	if req.About == "" {
		arg.About = user.About
	} else {
		arg.About = sql.NullString{
			String: req.About,
			Valid:  true,
		}
	}

	// edit user if the input values from query params are valid
	editedUser, err := s.DB.EditSellerByID(context.TODO(), arg)
	if err != nil {
		log.Warn("error updating users via EditSellerByID in EditProfileHandler for seller:", err.Error())
		http.Error(w, "internal server error updating seller profile values", http.StatusInternalServerError)
		return
	}

	// send response
	type respSeller struct {
		ID    uuid.UUID `json:"id"`
		Name  string    `json:"name"`
		About string    `json:"about"`
		Phone int       `json:"phone"`
		Email string    `json:"email"`
		GstNo string    `json:"gst_no"`
	}
	var respSellerData respSeller
	respSellerData.ID = editedUser.ID
	respSellerData.Name = editedUser.Name
	respSellerData.About = editedUser.About.String
	respSellerData.GstNo = editedUser.GstNo.String
	respSellerData.Phone = int(editedUser.Phone.Int64)
	respSellerData.Email = editedUser.Email

	var resp struct {
		Data    respSeller `json:"data"`
		Message string     `json:"message"`
	}
	resp.Data = respSellerData
	resp.Message = "successfully updated seller details"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Seller) OwnProductsHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	var Err []string
	products, err := s.DB.GetProductsBySellerID(context.TODO(), user.ID)
	if err == sql.ErrNoRows {
		Err = append(Err, "no products available for the seller yet.")
	} else if err != nil {
		log.Warn("error fetching products for seller: ", user.ID, ":", err)
		http.Error(w, "unable to fetch seller products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var resp struct {
		Data []db.Product `json:"data"`
		Err  []string     `json:"errors"`
	}
	resp.Data = products
	resp.Err = Err
	json.NewEncoder(w).Encode(resp)
}

func (s *Seller) ProductDetailsHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductID uuid.UUID `json:"product_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}
	product, err := s.DB.GetProductByID(context.TODO(), req.ProductID)
	if err == sql.ErrNoRows {
		http.Error(w, "not a valid productID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error fetching product from seller", err.Error())
		http.Error(w, "error fetching product", http.StatusInternalServerError)
		return
	}

	var Err []string
	categories, err := s.DB.GetCategoryNamesOfProductByID(context.TODO(), product.ID)
	if err == sql.ErrNoRows {
		Err = append(Err, "no categories added for product yet.")
	} else if err != nil {
		Err = append(Err, "error fetching categories for product")
	}
	var resp struct {
		Data       db.Product `json:"data"`
		Message    string     `json:"message"`
		Categories []string   `json:"categories"`
		Err        []string   `json:"errors"`
	}
	resp.Data = product
	resp.Categories = categories
	resp.Err = Err
	resp.Message = "successfully fetched product"
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Warn("error parsing response")
		http.Error(w, "error parsing response", http.StatusInternalServerError)
		return
	}
}

func (s *Seller) AddProductHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	// check if the seller added a shipping address or not
	_, err := s.DB.GetAddressBySellerID(context.TODO(), user.ID)
	if err == sql.ErrNoRows {
		http.Error(w, "cannot add product withot a address for seller. visit /seller/address/add and make an address", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error fetching address for seller in AddProductHandler:", err.Error())
		http.Error(w, "internal error fetching seller address to verify the seller has an address before adding product", http.StatusInternalServerError)
		return
	}
	var arg struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Price       float64  `json:"price"`
		Stock       int      `json:"stock"`
		Categories  []string `json:"categories"`
	}
	err = json.NewDecoder(r.Body).Decode(&arg)
	if err != nil {
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}
	if !(validators.ValidateProductName(arg.Name) &&
		validators.ValidateProductPrice(arg.Price) &&
		validators.ValidateProductStock(arg.Stock)) {
		http.Error(w, "invalid data values", http.StatusBadRequest)
		return
	}
	var productArg db.AddProductParams
	productArg.SellerID = user.ID
	productArg.Name = arg.Name
	productArg.Description = arg.Description
	productArg.Price = arg.Price
	productArg.Stock = int32(arg.Stock)
	product, err := s.DB.AddProduct(context.TODO(), productArg)
	if err != nil {
		log.Warnf("error adding product from sellerID: %s", user.ID)
		log.Warn(err)
		http.Error(w, "internal error while adding product", http.StatusInternalServerError)
		return
	}
	var Err []string
	var CategoriesAdded []string
	for _, v := range arg.Categories {
		var catArg db.AddProductToCategoryByCategoryNameParams
		catArg.CategoryName = v
		catArg.ProductID = product.ID
		_, err = s.DB.AddProductToCategoryByCategoryName(context.TODO(), catArg)
		if err != nil {
			Err = append(Err, "error adding product to category:"+v)
		} else {
			CategoriesAdded = append(CategoriesAdded, v)
		}
	}
	var resp struct {
		Data            db.Product `json:"data"`
		Message         string     `json:"message"`
		CategoriesAdded []string   `json:"categories_added"`
		Err             []string   `json:"error"`
	}
	resp.Data = product
	resp.Message = "product added successfully"
	resp.CategoriesAdded = CategoriesAdded
	resp.Err = Err
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Seller) EditProductHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	var req struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Price       float64   `json:"price"`
		Stock       int       `json:"stock"`
		Categories  []string  `json:"categories"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "wrong request format", http.StatusBadRequest)
		return
	} else if !(validators.ValidateProductName(req.Name) &&
		validators.ValidateProductPrice(req.Price) &&
		validators.ValidateProductStock(req.Stock)) {
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}

	var productID = req.ID
	seller, err := s.DB.GetSellerByProductID(context.TODO(), productID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Warn("error fetching sellerID from database")
		http.Error(w, "internal server error adding the product; database error", http.StatusInternalServerError)
		return
	} else if seller.ID != user.ID {
		http.Error(w, "trying to edit products not owned by you", http.StatusBadRequest)
		return
	}

	// logic
	var arg db.EditProductByIDParams
	arg.ID = req.ID
	arg.Name = req.Name
	arg.Description = req.Description
	arg.Price = req.Price
	arg.Stock = int32(req.Stock)
	product, err := s.DB.EditProductByID(context.TODO(), arg)
	if err == sql.ErrNoRows {
		http.Error(w, "no product with the specified id", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn(err)
		http.Error(w, "error updating product details", http.StatusInternalServerError)
		return
	}

	var Err []string
	var CategoriesAdded []string
	err = s.DB.DeleteAllCategoriesForProductByID(context.TODO(), productID)
	if err != nil {
		Err = append(Err, "error removing all the previous categories attached to the product:"+product.Name)
		log.Warn("error removing all the previous categories attached to the product on DeleteAllCategoriesForProductByID:", err.Error())
	}
	for _, v := range req.Categories {
		var CatArg db.AddProductToCategoryByCategoryNameParams
		CatArg.ProductID = productID
		CatArg.CategoryName = v
		_, err = s.DB.AddProductToCategoryByCategoryName(context.TODO(), CatArg)
		if err != nil {
			Err = append(Err, "error adding product to category:"+v)
		} else {
			CategoriesAdded = append(CategoriesAdded, v)
		}
	}

	var resp struct {
		Data            db.Product `json:"data"`
		Message         string     `json:"message"`
		Err             []string   `json:"error"`
		CategoriesAdded []string   `json:"categories_added"`
	}
	resp.Err = Err
	resp.CategoriesAdded = CategoriesAdded
	resp.Data = product
	resp.Message = "updated product details"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Seller) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	// take user from r context written by AuthenticateUserMiddleware
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	// arguemnt struct to unmarshall from r.Body
	var req struct {
		ProductID uuid.UUID `json:"product_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Warn(err)
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}

	// checking if the user.ID is the same as the product.SellerID
	var productID = req.ProductID
	seller, err := s.DB.GetSellerByProductID(context.TODO(), productID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid product_id", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error fetching sellerID from database")
		http.Error(w, "internal server error adding the product; database error", http.StatusInternalServerError)
		return
	} else if seller.ID != user.ID {
		http.Error(w, "trying to edit products not owned by you", http.StatusBadRequest)
		return
	}

	// business logic
	product, err := s.DB.DeleteProductByID(context.TODO(), req.ProductID)
	if err != nil {
		log.Warn(err)
		http.Error(w, "error deleting product", http.StatusInternalServerError)
		return
	}
	var resp struct {
		Product db.Product `json:"product"`
		Message string     `json:"message"`
	}
	resp.Product = product
	resp.Message = "successfully deleted product"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Seller) GetAllCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := s.DB.GetAllCategories(context.TODO())
	if err != nil {
		log.Warn("error fetching all categories for seller:", err)
		http.Error(w, "internal error fetching cateogries", http.StatusInternalServerError)
		return
	}
	var resp struct {
		Data    []db.Category `json:"data"`
		Message string        `json:"message"`
	}
	resp.Data = categories
	resp.Message = "successfully fetched all categories"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Seller) AddProductToCategoryHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("unable to fetch user from the request context after passing it from Auth Middleware")
		http.Error(w, "internal error fetching user by sessionID", http.StatusInternalServerError)
		return
	}
	var req struct {
		ProductID    uuid.UUID `json:"product_id"`
		CategoryName string    `json:"category_name"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "wrong request body format", http.StatusBadRequest)
		return
	}
	product, err := s.DB.GetProductByID(context.TODO(), req.ProductID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error fetching product from productID:", err.Error())
		http.Error(w, "internal server error fetching product from productID", http.StatusInternalServerError)
		return
	} else if product.SellerID != user.ID {
		http.Error(w, "user not authorized to add product to category, not user's product", http.StatusUnauthorized)
		return
	}
	category, err := s.DB.GetCategoryByName(context.TODO(), req.CategoryName)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid category", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error fetching category by name query")
		http.Error(w, "internal error fetching category", http.StatusInternalServerError)
		return
	}

	var arg db.AddProductToCategoryByCategoryNameParams
	arg.CategoryName = req.CategoryName
	arg.ProductID = req.ProductID
	_, err = s.DB.AddProductToCategoryByCategoryName(context.TODO(), arg)
	if err != nil {
		log.Warn("error adding product to category items")
		http.Error(w, "internal error adding product to category items"+err.Error(), http.StatusInternalServerError)
		return
	}
	var resp struct {
		Product      db.Product `json:"product"`
		CategoryName string     `json:"category_name"`
		Message      string     `json:"message"`
	}
	resp.Product = product
	resp.CategoryName = category.Name
	resp.Message = "successfully added product to category items"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Seller) GetAddressesHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	helper.GetAddressesHelper(w, r, user)
}
func (s *Seller) AddAddressHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	helper.AddAddressHelper(w, r, user)
}

func (s *Seller) EditAddressHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	helper.EditAddressHelper(w, r, user)
}

// /////////////////////////////////
// order handler

func (s *Seller) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	orderItems, err := s.DB.GetOrderItemsBySellerID(context.TODO(), user.ID)
	if err != nil {
		log.Warn("error fetching orderItems for the seller in GetOrdersHandler:", err.Error())
		http.Error(w, "intenral error fetching orderItems for the seller", http.StatusInternalServerError)
		return
	}

	// respOrderItem struct
	type respOrderItem struct {
		OrderItemID uuid.UUID `json:"order_item_id"`
		Status      string    `json:"status"`
		ProductID   uuid.UUID `json:"product_id"`
		Price       float64   `json:"price"`
		Quantity    int       `json:"quantity"`
		TotalAmount float64   `json:"total_amount"`
	}

	// respOrderItems slice for seller
	var respOrderItems []respOrderItem

	for _, v := range orderItems {
		var temp respOrderItem
		temp.OrderItemID = v.ID
		temp.Status = v.Status
		temp.ProductID = v.ProductID
		temp.Price = v.Price
		temp.Quantity = int(v.Quantity)
		temp.TotalAmount = v.TotalAmount

		respOrderItems = append(respOrderItems, temp)
	}

	// send response
	var resp struct {
		Data    []respOrderItem `json:"data"`
		Message string          `json:"message"`
	}
	resp.Data = respOrderItems
	resp.Message = "successfully fetched seller's orderItems"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Seller) ChangeOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	var req struct {
		OrderItemIDStr string `json:"order_item_id"`
		Status         string `json:"status"`
	}
	req.OrderItemIDStr = r.URL.Query().Get("order_item_id")
	req.Status = r.URL.Query().Get("status")
	// get orderItemID
	orderItemID, err := uuid.Parse(req.OrderItemIDStr)
	if err != nil {
		http.Error(w, "invalid ordreItemID", http.StatusBadRequest)
		return
	}
	// get OrderItem
	orderItem, err := s.DB.GetOrderItemByID(context.TODO(), orderItemID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid orderItemID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error fetching orderItem in ChangeOrderItemStatusHandler in seller:", err.Error())
		http.Error(w, "internal error fetching orderItem", http.StatusInternalServerError)
		return
	}
	sellerID, err := s.DB.GetSellerIDFromOrderItemID(context.TODO(), orderItem.ID)
	if err != nil {
		log.Warn("error fetching sellerID from orderItemID in ChangeOrderStatusHandler in seller:", err.Error())
		http.Error(w, "internal error fetching sellerID from orderItem to verify it is the same seller's id", http.StatusInternalServerError)
		return
	}
	if sellerID != user.ID {
		http.Error(w, "not the current sellers's order_item to change status", http.StatusUnauthorized)
		return
	}
	// check if orderStatus, req status is valid for update
	if orderItem.Status == utils.StatusOrderCancelled ||
		orderItem.Status == utils.StatusOrderDelivered ||
		orderItem.Status == utils.StatusOrderShipped {
		msg := fmt.Sprintf("orderItem %s. cannot change the status of ordreItem", orderItem.Status)
		http.Error(w, msg, http.StatusBadRequest)
	} else if req.Status != utils.StatusOrderPending &&
		req.Status != utils.StatusOrderProcessing &&
		req.Status != utils.StatusOrderShipped {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	// udpate orderItemStatus
	var arg db.ChangeOrderItemStatusByIDParams
	arg.ID = orderItem.ID
	arg.Status = req.Status
	updatedOrderItem, err := s.DB.ChangeOrderItemStatusByID(context.TODO(), arg)
	if err != nil {
		log.Warn("error updating ordreItem status for seller ChangeOrderStatusHandler:", err.Error())
		http.Error(w, "internal error changing status for orderItem", http.StatusInternalServerError)
		return
	}

	// send response
	type respOrderItem struct {
		OrderItemID uuid.UUID `json:"order_item_id"`
		Status      string    `json:"status"`
		ProductID   uuid.UUID `json:"product_id"`
		Price       float64   `json:"price"`
		Quantity    int       `json:"quantity"`
		TotalAmount float64   `json:"total_amount"`
	}

	var updatedRespOrderItem respOrderItem
	updatedRespOrderItem.OrderItemID = updatedOrderItem.ID
	updatedRespOrderItem.Status = updatedOrderItem.Status
	updatedRespOrderItem.ProductID = updatedOrderItem.ProductID
	updatedRespOrderItem.Price = updatedOrderItem.Price
	updatedRespOrderItem.Quantity = int(updatedOrderItem.Quantity)
	updatedRespOrderItem.TotalAmount = updatedOrderItem.TotalAmount

	var resp struct {
		Data    respOrderItem `json:"data"`
		Message string        `json:"message"`
	}

	resp.Data = updatedRespOrderItem
	resp.Message = "successfully updated orderItem status"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Seller) SalesReportHandler(w http.ResponseWriter, r *http.Request) {
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

	orderItems, err := s.DB.GetOrderItemsBySellerIDAndDateRange(context.TODO(), db.GetOrderItemsBySellerIDAndDateRangeParams{
		SellerID:  user.ID,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		log.Println("Error fetching order items:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	vendorPayments, err := s.DB.GetVendorPaymentsBySellerIDAndDateRange(context.TODO(), db.GetVendorPaymentsBySellerIDAndDateRangeParams{
		SellerID:  user.ID,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		log.Println("Error fetching vendor payments:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Aggregated calculations
	totalSales := 0.0
	totalCredit := 0.0
	totalPlatformFee := 0.0
	orderCancelledCount := 0
	orderWaitingCount := 0
	orderPendingCount := 0
	productSales := make(map[uuid.UUID]map[string]float64)

	for _, payment := range vendorPayments {
		if payment.Status == utils.StatusVendorPaymentCancelled {
			orderCancelledCount += 1
			continue
		} else if payment.Status == utils.StatusVendorPaymentWaiting {
			orderWaitingCount += 1
			continue
		} else if payment.Status == utils.StatusVendorPaymentPending {
			orderPendingCount += 1
			continue
		}
		totalSales += payment.TotalAmount
		totalCredit += payment.CreditAmount
		totalPlatformFee += payment.PlatformFee

		if _, exists := productSales[payment.OrderItemID]; !exists {
			productSales[payment.OrderItemID] = map[string]float64{"sales": 0, "profit": 0, "platform_fee": 0}
		}
		productSales[payment.OrderItemID]["sales"] += payment.TotalAmount
		productSales[payment.OrderItemID]["profit"] += payment.CreditAmount
		productSales[payment.OrderItemID]["platform_fee"] += payment.PlatformFee
	}

	// Generate PDF report
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Sales Report")
	pdf.Ln(10)

	// Date range
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(95, 10, fmt.Sprintf("Start Date: %s", startDate.Format("2006-01-02")))
	pdf.Cell(95, 10, fmt.Sprintf("End Date: %s", endDate.Format("2006-01-02")))
	pdf.Ln(10)

	// Summary
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "Summary")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(95, 8, fmt.Sprintf("Total Sales: $%.2f", totalSales))
	pdf.Cell(95, 8, fmt.Sprintf("Total Profit: $%.2f", totalCredit))
	pdf.Ln(8)
	pdf.Cell(95, 8, fmt.Sprintf("Total Platform Fee: $%.2f", totalPlatformFee))
	pdf.Ln(10)

	// Table Headers
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(60, 8, "Product ID", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Sales ($)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Profit ($)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Platform Fee ($)", "1", 1, "C", true, 0, "")

	// Table Content
	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(240, 240, 240)
	fill := false

	for productID, data := range productSales {
		pdf.CellFormat(60, 8, productID.String(), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", data["sales"]), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", data["profit"]), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", data["platform_fee"]), "1", 1, "C", fill, 0, "")
		fill = !fill
	}

	// Check if order items exist
	if len(orderItems) > 0 {
		pdf.Ln(10)
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 8, "Order Items Details")
		pdf.Ln(8)

		// Table Headers for Order Items
		pdf.SetFont("Arial", "B", 12)
		pdf.SetFillColor(200, 200, 200)
		pdf.CellFormat(40, 8, "Order ID", "1", 0, "C", true, 0, "")
		pdf.CellFormat(60, 8, "Product ID", "1", 0, "C", true, 0, "")
		pdf.CellFormat(40, 8, "Quantity", "1", 0, "C", true, 0, "")
		pdf.CellFormat(40, 8, "Price ($)", "1", 1, "C", true, 0, "")

		// Table Content
		pdf.SetFont("Arial", "", 10)
		pdf.SetFillColor(240, 240, 240)
		fill = false

		for _, order := range orderItems {
			pdf.CellFormat(40, 8, order.OrderID.String(), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(60, 8, order.ProductID.String(), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(40, 8, fmt.Sprintf("%d", order.Quantity), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", order.Price), "1", 1, "C", fill, 0, "")
			fill = !fill
		}
	}

	// Output PDF
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		http.Error(w, "Error generating PDF", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=sales_report.pdf")
	w.Write(buf.Bytes())
}
