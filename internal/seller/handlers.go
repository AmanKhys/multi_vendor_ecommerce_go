package seller

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/chartGen.go"
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

// get Seller profile Handler
func (u *Seller) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	wallet, err := u.DB.GetWalletByUserID(context.TODO(), user.ID)
	if err == sql.ErrNoRows {
		log.Error("error no wallet assinged to user; error in GetPRofileHandler for seller:", err.Error())
		http.Error(w, "internal error: no wallet assigned to user", http.StatusInternalServerError)
		return
	} else if err != nil {
		log.Error("error fetching wallet for seller in GetProfileHandler for seller:", err.Error())
		http.Error(w, "internal error: unable to fetch necessary data", http.StatusInternalServerError)
		return
	}

	address, err := u.DB.GetAddressBySellerID(context.TODO(), user.ID)
	if err != nil {
		log.Error("error fetching address for seller in GetProfileHandler:", err.Error())
		http.Error(w, "internal error: unable to fetch necessary data", http.StatusInternalServerError)
		return
	}

	type addressResp struct {
		ID           uuid.UUID `json:"id"`
		BuildingName string    `json:"building_name"`
		StreetName   string    `json:"street_name"`
		Town         string    `json:"town"`
		District     string    `json:"district"`
		State        string    `json:"state"`
		Pincode      int32     `json:"pincode"`
	}

	var respAddress addressResp
	respAddress.ID = address.ID
	respAddress.BuildingName = address.BuildingName
	respAddress.StreetName = address.StreetName
	respAddress.Town = address.Town
	respAddress.District = address.District
	respAddress.Pincode = address.Pincode
	respAddress.State = address.State

	var resp struct {
		ID      uuid.UUID   `json:"user_id"`
		Name    string      `json:"name"`
		About   string      `json:"about"`
		Phone   int         `json:"phone"`
		Email   string      `json:"email"`
		Wallet  float64     `json:"wallet_savings"`
		Address addressResp `json:"addresses"`
	}
	resp.ID = user.ID
	resp.Name = user.Name
	resp.About = user.About.String
	resp.Phone = int(user.Phone.Int64)
	resp.Email = user.Email
	resp.Wallet = wallet.Savings
	resp.Address = respAddress
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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
	productIDStr := r.URL.Query().Get("product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		http.Error(w, "not a valid product_id", http.StatusBadRequest)
		return
	}
	product, err := s.DB.GetProductByID(context.TODO(), productID)
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
	timeLimit := r.URL.Query().Get("time_limit")

	var orderItems []db.OrderItem
	var startDate time.Time
	var endDate time.Time

	if timeLimit == "day" {
		startDate = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
		endDate = time.Now()
	} else if timeLimit == "week" {
		startDate = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
		day := int(time.Now().Weekday())
		realStartDate := startDate.AddDate(0, 0, -day)
		startDate = realStartDate
		endDate = time.Now()
	} else if timeLimit == "month" {
		startDate = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
		endDate = time.Now()
	} else if timeLimit == "year" {
		startDate = time.Date(time.Now().Year(), time.January, 1, 0, 0, 0, 0, time.Now().Location())
		endDate = time.Now()
	} else {
		sd, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		ed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		startDate = sd
		endDate = ed
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

	var productOrders = make(map[uuid.UUID]int)
	for _, v := range orderItems {
		productOrders[v.ProductID] += int(v.Quantity)
	}

	// Convert map to slice for sorting
	type productStat struct {
		ProductID   uuid.UUID `json:"product_id"`
		ProductName string    `json:"product_name"`
		Quantity    int       `json:"quantity"`
	}
	var sortedProducts []productStat
	for pid, qty := range productOrders {
		product, err := s.DB.GetProductByID(context.TODO(), pid)
		if err != nil {
			log.Error("error fetching product by productID in SalesReportHandler for seller")
			continue
		}
		sortedProducts = append(sortedProducts, productStat{ProductID: pid, ProductName: product.Name, Quantity: qty})
	}

	// Sort by quantity in descending order
	sort.Slice(sortedProducts, func(i, j int) bool {
		return sortedProducts[i].Quantity > sortedProducts[j].Quantity
	})

	// Select top 3
	topThree := sortedProducts
	if len(sortedProducts) > 3 {
		topThree = sortedProducts[:3]
	}

	// get a table of the details of each orderItem
	orderStatusCounts := map[string]int{
		"Pending":    0,
		"Processing": 0,
		"Shipped":    0,
		"Delivered":  0,
		"Returned":   0,
		"Cancelled":  0,
	}
	for _, oi := range orderItems {
		switch oi.Status {
		case utils.StatusOrderPending:
			orderStatusCounts["Pending"]++
		case utils.StatusOrderProcessing:
			orderStatusCounts["Processing"]++
		case utils.StatusOrderShipped:
			orderStatusCounts["Shipped"]++
		case utils.StatusOrderDelivered:
			orderStatusCounts["Delivered"]++
		case utils.StatusOrderReturned:
			orderStatusCounts["Returned"]++
		case utils.StatusOrderCancelled:
			orderStatusCounts["Cancelled"]++
		}

	}

	paymentStatusCounts := map[string]chartGen.PaymentStat{
		"Pending":   {Count: 0, Amount: 0},
		"Waiting":   {Count: 0, Amount: 0},
		"Failed":    {Count: 0, Amount: 0},
		"Received":  {Count: 0, Amount: 0},
		"Cancelled": {Count: 0, Amount: 0},
	}

	var platformFees float64
	for _, vp := range vendorPayments {
		switch vp.Status {
		case utils.StatusVendorPaymentPending:
			entry := paymentStatusCounts["Pending"]
			entry.Count++
			entry.Amount += vp.CreditAmount
			paymentStatusCounts["Pending"] = entry
		case utils.StatusVendorPaymentWaiting:
			entry := paymentStatusCounts["Waiting"]
			entry.Count++
			entry.Amount += vp.CreditAmount
			paymentStatusCounts["Waiting"] = entry
		case utils.StatusVendorPaymentFailed:
			entry := paymentStatusCounts["Failed"]
			entry.Count++
			entry.Amount += vp.CreditAmount
			paymentStatusCounts["Failed"] = entry
		case utils.StatusVendorPaymentReceived:
			entry := paymentStatusCounts["Received"]
			entry.Count++
			entry.Amount += vp.CreditAmount
			paymentStatusCounts["Received"] = entry
			platformFees += vp.PlatformFee
		case utils.StatusVendorPaymentCancelled:
			entry := paymentStatusCounts["Cancelled"]
			entry.Count++
			entry.Amount += vp.CreditAmount
			paymentStatusCounts["Cancelled"] = entry
		}

	}

	netProfit := paymentStatusCounts["Received"].Amount - platformFees

	pieChartPath, barChartPath, err := chartGen.GenerateChartsForSeller(orderStatusCounts, paymentStatusCounts, platformFees, netProfit)
	if err != nil {
		log.Println("Failed to generate charts:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer os.Remove(pieChartPath)
	defer os.Remove(barChartPath)

	pdf := gofpdf.New("P", "mm", "A3", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	// Main title
	pdf.SetFont("Arial", "B", 18)
	pdf.CellFormat(0, 12, "Seller Sales Report", "", 1, "C", false, 0, "")
	pdf.Ln(2)

	// Report Date Range
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(95, 8, fmt.Sprintf("Start Date: %s", startDate.Format("02 Jan 2006")))
	pdf.Cell(95, 8, fmt.Sprintf("End Date: %s", endDate.Format("02 Jan 2006")))
	pdf.Ln(12)

	// ========== TOP 3 SELLING PRODUCTS ========== //
	pdf.SetFont("Arial", "B", 13)
	pdf.Cell(0, 8, "Top 3 Selling Products")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(230, 230, 230)
	pdf.CellFormat(10, 8, "No.", "1", 0, "C", true, 0, "")
	pdf.CellFormat(100, 8, "Product Name", "1", 0, "", true, 0, "")
	pdf.CellFormat(80, 8, "Product ID", "1", 0, "", true, 0, "")
	pdf.CellFormat(30, 8, "Quantity Sold", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	for i, p := range topThree {
		pdf.CellFormat(10, 8, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(100, 8, p.ProductName, "1", 0, "", false, 0, "")
		pdf.CellFormat(80, 8, p.ProductID.String(), "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 8, fmt.Sprintf("%d", p.Quantity), "1", 1, "C", false, 0, "")
	}
	pdf.Ln(12)

	// ========== ORDER ITEMS SECTION ========== //
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Product Sales Report", "", 1, "", false, 0, "")
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(10, pdf.GetY(), 290, pdf.GetY())
	pdf.Ln(6)

	// Order Items Summary
	pdf.SetFont("Arial", "B", 13)
	pdf.Cell(0, 8, "Order Items Summary")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	linBreakFlag := false
	for k, v := range orderStatusCounts {
		pdf.Cell(95, 8, fmt.Sprintf("%-10s : %d", k, v))
		if linBreakFlag {
			pdf.Ln(8)
		}
		linBreakFlag = !linBreakFlag
	}
	pdf.Ln(12)

	// Table Header
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(230, 230, 230)
	pdf.CellFormat(10, 10, "No.", "1", 0, "C", true, 0, "")
	pdf.CellFormat(80, 10, "OrderItem ID", "1", 0, "", true, 0, "")
	pdf.CellFormat(80, 10, "Product ID", "1", 0, "", true, 0, "")
	pdf.CellFormat(25, 10, "Status", "1", 0, "", true, 0, "")
	pdf.CellFormat(25, 10, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 10, "Price", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	slno := 1
	var fillFlag bool
	for _, oi := range orderItems {
		pdf.CellFormat(10, 8, fmt.Sprintf("%d", slno), "1", 0, "C", fillFlag, 0, "")
		pdf.CellFormat(80, 8, oi.ID.String(), "1", 0, "", !fillFlag, 0, "")
		pdf.CellFormat(80, 8, oi.ProductID.String(), "1", 0, "", fillFlag, 0, "")
		pdf.CellFormat(25, 8, oi.Status, "1", 0, "", !fillFlag, 0, "")
		pdf.CellFormat(25, 8, fmt.Sprintf("%d", oi.Quantity), "1", 0, "C", fillFlag, 0, "")
		pdf.CellFormat(25, 8, fmt.Sprintf("%.2f", oi.Price), "1", 1, "R", !fillFlag, 0, "")
		slno++
		fillFlag = !fillFlag
	}

	// ========== PAYMENT SECTION ========== //
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Payments Report", "", 1, "", false, 0, "")
	pdf.Line(10, pdf.GetY(), 290, pdf.GetY())
	pdf.Ln(6)

	// Payment Summary
	pdf.SetFont("Arial", "B", 13)
	pdf.Cell(0, 8, "Vendor Payments Summary")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	linBreakFlag = false
	for k, v := range paymentStatusCounts {
		pdf.Cell(95, 8, fmt.Sprintf("%-10s : %d ($%.2f)", k, v.Count, v.Amount))
		if linBreakFlag {
			pdf.Ln(8)
		}
		linBreakFlag = !linBreakFlag
	}
	pdf.Ln(8)
	pdf.Cell(0, 8, fmt.Sprintf("Debited Platform Fee: $%.2f", platformFees))
	pdf.Ln(6)
	pdf.Cell(0, 8, fmt.Sprintf("Net Profit: $%.2f", netProfit))
	pdf.Ln(12)

	// Vendor Payments Table
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(230, 230, 230)
	pdf.CellFormat(10, 8, "No.", "1", 0, "C", true, 0, "")
	pdf.CellFormat(80, 8, "Order Item ID", "1", 0, "", true, 0, "")
	pdf.CellFormat(80, 8, "Vendor Payment ID", "1", 0, "", true, 0, "")
	pdf.CellFormat(25, 8, "Status", "1", 0, "", true, 0, "")
	pdf.CellFormat(25, 8, "Total", "1", 0, "R", true, 0, "")
	pdf.CellFormat(25, 8, "Credit", "1", 0, "R", true, 0, "")
	pdf.CellFormat(25, 8, "Ptf Fee", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	slno = 1
	fillFlag = false
	for _, vp := range vendorPayments {
		pdf.CellFormat(10, 8, fmt.Sprintf("%d", slno), "1", 0, "C", fillFlag, 0, "")
		pdf.CellFormat(80, 8, vp.OrderItemID.String(), "1", 0, "", !fillFlag, 0, "")
		pdf.CellFormat(80, 8, vp.ID.String(), "1", 0, "", fillFlag, 0, "")
		pdf.CellFormat(25, 8, vp.Status, "1", 0, "", !fillFlag, 0, "")
		pdf.CellFormat(25, 8, fmt.Sprintf("%.2f", vp.TotalAmount), "1", 0, "R", fillFlag, 0, "")
		pdf.CellFormat(25, 8, fmt.Sprintf("%.2f", vp.CreditAmount), "1", 0, "R", !fillFlag, 0, "")
		pdf.CellFormat(25, 8, fmt.Sprintf("%.2f", vp.PlatformFee), "1", 1, "R", fillFlag, 0, "")
		slno++
		fillFlag = !fillFlag
	}

	// ========== CHARTS ========== //
	pdf.AddPage()
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Visual Summary")
	pdf.Ln(10)
	pdf.Image(pieChartPath, 20, pdf.GetY(), 120, 0, false, "", 0, "")
	pdf.Image(barChartPath, 150, pdf.GetY(), 120, 0, false, "", 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		log.Println("Error generating PDF:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=sales_report.pdf")
	w.Write(buf.Bytes())
}
