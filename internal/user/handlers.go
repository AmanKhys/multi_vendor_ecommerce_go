package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/envname"
	helpers "github.com/amankhys/multi_vendor_ecommerce_go/pkg/payment"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/validators"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"github.com/google/uuid"
	env "github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type User struct {
	DB *db.Queries
}

// get User profile Handler
func (u *User) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	wallet, err := u.DB.GetWalletByUserID(context.TODO(), user.ID)
	if err == sql.ErrNoRows {
		log.Error("error no wallet assinged to user; error in GetPRofileHandler for user:", err.Error())
		http.Error(w, "internal error: no wallet assigned to user", http.StatusInternalServerError)
		return
	} else if err != nil {
		log.Error("error fetching wallet for user in GetProfileHandler for user:", err.Error())
		http.Error(w, "internal error: unable to fetch necessary data", http.StatusInternalServerError)
		return
	}

	addresses, err := u.DB.GetAddressesByUserID(context.TODO(), user.ID)
	if err != nil {
		log.Error("error fetching addresses for user in GetProfileHandler:", err.Error())
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

	var addressesResp []addressResp
	for _, v := range addresses {
		var temp addressResp
		temp.ID = v.ID
		temp.BuildingName = v.BuildingName
		temp.StreetName = v.StreetName
		temp.Town = v.Town
		temp.District = v.District
		temp.Pincode = v.Pincode
		temp.State = v.State

		addressesResp = append(addressesResp, temp)
	}

	var resp struct {
		ID        uuid.UUID     `json:"user_id"`
		Name      string        `json:"name"`
		Phone     sql.NullInt64 `json:"phone"`
		Email     string        `json:"email"`
		Wallet    float64       `json:"wallet_savings"`
		Addresses []addressResp `json:"addresses"`
	}
	resp.ID = user.ID
	resp.Name = user.Name
	resp.Phone = user.Phone
	resp.Email = user.Email
	resp.Wallet = wallet.Savings
	resp.Addresses = addressesResp
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ///////////////////////////////////
// edit profile handler
func (u *User) EditProfileHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	// get request
	var req struct {
		Name  string
		Phone string
	}
	req.Name = r.URL.Query().Get("name")
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
	var arg db.EditUserByIDParams
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

	// edit user if the input values from query params are valid
	editedUser, err := u.DB.EditUserByID(context.TODO(), arg)
	if err != nil {
		log.Warn("error updating users via EditUserByID in EditProfileHandler for seller:", err.Error())
		http.Error(w, "internal server error updating seller profile values", http.StatusInternalServerError)
		return
	}

	// send response
	type respUser struct {
		ID    uuid.UUID `json:"id"`
		Name  string    `json:"name"`
		Phone int       `json:"phone"`
		Email string    `json:"email"`
	}
	var respUserData respUser
	respUserData.ID = editedUser.ID
	respUserData.Name = editedUser.Name
	respUserData.Phone = int(editedUser.Phone.Int64)
	respUserData.Email = editedUser.Email

	var resp struct {
		Data    respUser `json:"data"`
		Message string   `json:"message"`
	}
	resp.Data = respUserData
	resp.Message = "successfully updated seller details"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// //////////////////////////////////
// product handlers

func (u *User) ProductsHandler(w http.ResponseWriter, r *http.Request) {
	// take request value params
	var req struct {
		PriceMaxStr string   `json:"price_max"`
		PriceMinStr string   `json:"price_min"`
		Categories  []string `json:"categories"`
		Name        string   `json:"name"`
	}
	req.PriceMaxStr = r.URL.Query().Get("price_max")
	req.PriceMinStr = r.URL.Query().Get("price_min")
	req.Categories = r.URL.Query()["categories"]
	req.Name = r.URL.Query().Get("name")

	// take valid values out of request
	PriceMax, err := strconv.Atoi(req.PriceMaxStr)
	if err != nil || PriceMax < 0 {
		PriceMax = math.MaxInt
	}
	PriceMin, err := strconv.Atoi(req.PriceMinStr)
	if err != nil || PriceMin < 0 {
		PriceMin = 0
	}
	var Categories []string
	for _, v := range req.Categories {
		category, err := u.DB.GetCategoryByName(context.TODO(), v)
		if err == sql.ErrNoRows {
			continue
		} else if err != nil {
			log.Warn("error fetching categoryByName in ProductsHandler in user:", err.Error())
			continue
		}
		Categories = append(Categories, category.Name)
	}
	// take all products
	allProducts, err := u.DB.GetAllProducts(context.TODO())
	if err != nil {
		log.Warn("error fetching products in ProductsHandler in user:", err.Error())
		http.Error(w, "internal server error fetching products", http.StatusInternalServerError)
		return
	}

	// name filter by either name or description
	var nameFilterProducts []db.Product
	if len(req.Name) > 0 {
		for _, v := range allProducts {
			if utils.FilterName(req.Name, v.Name+v.Description) {
				nameFilterProducts = append(nameFilterProducts, v)
			}
		}
	} else {
		nameFilterProducts = allProducts
	}

	// take category filtered products
	var filteredProducts []db.Product
	if len(Categories) > 0 {
		for _, v := range nameFilterProducts {
			vCats, err := u.DB.GetCategoryNamesOfProductByID(context.TODO(), v.ID)
			if err != nil {
				log.Warn("error fetching category names of product by id in ProductsHandler for user")
				continue
			} else if utils.CheckCategory(vCats, Categories) {
				filteredProducts = append(filteredProducts, v)
			}
		}
	} else {
		filteredProducts = nameFilterProducts
	}

	// filter products by price
	var finalProducts []db.Product
	for _, v := range filteredProducts {
		if v.Price <= float64(PriceMax) && v.Price >= float64(PriceMin) {
			finalProducts = append(finalProducts, v)
		}
	}

	// make response product struct
	type respProduct struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Price       float64   `json:"price"`
		Stock       int       `json:"stock"`
		SellerID    uuid.UUID `json:"seller_id"`
	}
	var respProducts []respProduct
	for _, v := range finalProducts {
		var temp respProduct
		temp.ID = v.ID
		temp.Name = v.Name
		temp.Description = v.Description
		temp.Price = v.Price
		temp.Stock = int(v.Stock)
		temp.SellerID = v.SellerID

		respProducts = append(respProducts, temp)
	}

	// send response
	var resp struct {
		Data    []respProduct `json:"data"`
		Message string        `json:"message"`
	}
	w.Header().Set("Content-Type", "application/json")
	resp.Data = respProducts
	resp.Message = "successfully fetched filtered products"
	json.NewEncoder(w).Encode(resp)
}

func (u *User) ProductHandler(w http.ResponseWriter, r *http.Request) {
	ProductID := r.URL.Query().Get("id")
	id, err := uuid.Parse(ProductID)
	if err != nil {
		http.Error(w, "wrong productID format", http.StatusBadRequest)
		return
	}

	var Err []string
	var Messages []string

	product, err := u.DB.GetProductByID(context.TODO(), id)
	if err != nil {
		http.Error(w, "no such product exists", http.StatusNotFound)
		return
	}
	reviews, err := u.DB.GetProductReviews(context.TODO(), product.ID)
	var averageRating float64
	var totalRating int
	type respReview struct {
		ReviewID uuid.UUID      `json:"review_id"`
		Rating   int            `json:"rating"`
		Comment  sql.NullString `json:"comment"`
	}
	var respReviews []respReview
	if err != nil {
		Messages = append(Messages, "no reveiws added for this product as of yet")
	} else {
		result, err := u.DB.GetProductAverageRatingAndTotalRating(context.TODO(), product.ID)
		if err != nil {
			Err = append(Err, "error fetching average rating for the product")
		}
		averageRating = result.AverageRating
		totalRating = int(result.TotalRating)

		for _, r := range reviews {
			var temp respReview
			temp.ReviewID = r.ID
			temp.Rating = int(r.Rating)
			temp.Comment = r.Comment
			respReviews = append(respReviews, temp)
		}
	}

	var resp struct {
		ProductID     uuid.UUID       `json:"product_id"`
		Name          string          `json:"name"`
		Price         float64         `json:"price"`
		AverageRating sql.NullFloat64 `json:"average_rating"`
		RatingCount   int             `json:"rating_count"`
		Reviews       []respReview    `json:"reviews"`
		Err           []string        `json:"errors"`
		Messages      []string        `json:"messages"`
	}
	resp.ProductID = product.ID
	resp.Name = product.Name
	resp.Price = product.Price
	if averageRating != 0 {
		resp.AverageRating.Float64 = averageRating
		resp.AverageRating.Valid = true
	}
	resp.RatingCount = totalRating
	resp.Reviews = respReviews
	resp.Err = Err
	resp.Messages = Messages

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) AddProductReviewHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	productIDStr := r.URL.Query().Get("product_id")
	ratingStr := r.URL.Query().Get("rating")
	comment := r.URL.Query().Get("comment")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		http.Error(w, "invalid productID", http.StatusBadRequest)
		return
	}

	review, err := u.DB.GetReviewByUserAndProductID(context.TODO(), db.GetReviewByUserAndProductIDParams{
		UserID:    user.ID,
		ProductID: productID,
	})
	if err == nil {
		http.Error(w, "user already added review. Kindly Edit review if further changes are to be done.", http.StatusBadRequest)
		return
	}
	_, err = u.DB.GetOrderItemByUserAndProductID(context.TODO(), db.GetOrderItemByUserAndProductIDParams{
		UserID:    user.ID,
		ProductID: productID,
	})
	if err == sql.ErrNoRows {
		http.Error(w, "cannot add rating to unpurchased item", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Error("error fetching orderItem in AddProductReviewHandler:", err.Error())
		http.Error(w, "internal error fetching necessary item to add review", http.StatusInternalServerError)
		return
	}

	if !validators.ValidateReviewRating(ratingStr) {
		http.Error(w, "Invalid review rating. Rate from 1-5", http.StatusBadRequest)
		return
	}
	rating, _ := strconv.Atoi(ratingStr)
	if len(comment) == 0 {
		review, err = u.DB.AddProductReviewWithoutComment(context.TODO(), db.AddProductReviewWithoutCommentParams{
			UserID:    user.ID,
			ProductID: productID,
			Rating:    int32(rating),
		})
		if err != nil {
			log.Error("error adding review in AddProductReviewHandler:", err.Error())
			http.Error(w, "internal error adding review", http.StatusInternalServerError)
			return
		}
	} else {
		review, err = u.DB.AddProductReviewWithCommment(context.TODO(), db.AddProductReviewWithCommmentParams{
			UserID:    user.ID,
			ProductID: productID,
			Rating:    int32(rating),
			Comment: sql.NullString{
				String: comment,
				Valid:  true,
			},
		})
		if err != nil {
			log.Error("error adding review in AddProductReviewHandler:", err.Error())
			http.Error(w, "internal error adding review", http.StatusInternalServerError)
			return
		}
	}

	var resp struct {
		ProductID uuid.UUID      `json:"product_id"`
		Rating    int            `json:"rating"`
		Comment   sql.NullString `json:"comment"`
		Message   string         `json:"message"`
		IsEdited  bool           `json:"is_edited"`
	}
	resp.ProductID = productID
	resp.Rating = int(review.Rating)
	resp.Comment = review.Comment
	resp.IsEdited = review.IsEdited
	resp.Message = "successfully added review to the product"
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) CategoryHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	categoryName := queryParams.Get("category_name")

	products, err := u.DB.GetProductsByCategoryName(r.Context(), categoryName)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "text/plain")
		message := "no available products by the cateogry:" + categoryName
		w.Write([]byte(message))
	} else if err != nil {
		log.Warn("error fetching products by category name:", err.Error())
		http.Error(w, "internal error fetching products by category name", http.StatusBadRequest)
		return
	}

	var resp struct {
		Data    []db.GetProductsByCategoryNameRow `json:"data"`
		Message string                            `json:"message"`
	}
	resp.Data = products
	resp.Message = "successfully fetched products from category:" + categoryName
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ///////////////////////////////
// address handlers

func (u *User) GetAddressesHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	helper.GetAddressesHelper(w, r, user)
}
func (u *User) AddAddressHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	helper.AddAddressHelper(w, r, user)
}

func (u *User) EditAddressHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	helper.EditAddressHelper(w, r, user)
}

func (u *User) DeleteAddressHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	helper.DeleteAddressHelper(w, r, user)
}

// /////////////////////////////////////
// cart handlers

func (u *User) GetCartHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	cartItems, err := u.DB.GetCartItemsByUserID(r.Context(), user.ID)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		message := "no cart items added for user"
		w.Write([]byte(message))
		return
	} else if err != nil {
		log.Warn("error fetching cart items:", err.Error())
		http.Error(w, "internal server error fetching cart items", http.StatusInternalServerError)
		return
	}

	var resp struct {
		Data    []db.GetCartItemsByUserIDRow `json:"data"`
		Message string                       `json:"message"`
	}
	resp.Data = cartItems
	resp.Message = "successfully fetched cart items"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) AddCartHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	// get request and validate request body and validate it
	// create Err slice to give the errors for response
	var Err []string
	var req struct {
		ProductID uuid.UUID `json:"product_id"`
		Quantity  int       `json:"quantity"`
	}
	type respCartItem struct {
		CartID      uuid.UUID `json:"cart_id"`
		ProductID   uuid.UUID `json:"product_id"`
		ProductName string    `json:"product_name"`
		Quantity    int       `json:"quantity"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request data format", http.StatusBadRequest)
		return
	} else if req.Quantity < 0 {
		Err = append(Err, "trying to add invalid quantity. reverting quantity to 1")
		req.Quantity = 1
	} else if req.Quantity > 20 {
		Err = append(Err, "trying to add invalid quantity. reverting quantity to maximum possible 20")
		req.Quantity = 20
	}

	product, err := u.DB.GetProductByID(context.TODO(), req.ProductID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid productID", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "internal server error fetching product", http.StatusInternalServerError)
		return
	} else if product.Stock == 0 {
		http.Error(w, "product out of stock. cannot add item to cart", http.StatusBadRequest)
		return
	} else if product.Stock < int32(req.Quantity) {
		Err = append(Err, "product quantity added more than stock. Reverting to the maximum available stock for order.")
		req.Quantity = int(product.Stock)
	}

	// get if there are any product with the same productID already in cart
	// if so, update the cart item instead of adding a new one
	// else add the product to the cart
	var getArg db.GetCartItemByUserIDAndProductIDParams
	getArg.ProductID = product.ID
	getArg.UserID = user.ID
	cartItem, err := u.DB.GetCartItemByUserIDAndProductID(context.TODO(), getArg)
	// add cartItem if carts doesn't already have the particular combination of cartItem
	if err == sql.ErrNoRows {
		var arg db.AddCartItemParams
		arg.UserID = user.ID
		arg.ProductID = req.ProductID
		if req.Quantity == 0 {
			http.Error(w, "trying to add a product with zero quantity. Skipping the product to add.", http.StatusBadRequest)
			return
		} else {
			arg.Quantity = int32(req.Quantity)
		}
		arg.ProductID = product.ID
		arg.UserID = user.ID
		item, err := u.DB.AddCartItem(context.TODO(), arg)
		if err != nil {
			log.Warn("internal error adding cartItem:", err.Error())
			http.Error(w, "internal error adding cartItem", http.StatusInternalServerError)
			return
		}
		var respItem = respCartItem{
			CartID:      item.ID,
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    int(item.Quantity),
		}
		var resp struct {
			Data    respCartItem `json:"data"`
			Message string       `json:"message"`
			Err     []string     `json:"errors"`
		}
		resp.Data = respItem
		resp.Message = "successfully added cart item"
		resp.Err = Err

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return // make sure to return after sending the response

		// handle error in case an internal error fetching the cart
	} else if err != nil {
		log.Error(w, "internal error fetching cartItem to check before AddCart:", err.Error())
		http.Error(w, "internal error fetching cartItem to check before AddCart", http.StatusInternalServerError)
		return
	}

	// update existing cartItem if there is a valid cartItem that is fetched from the database
	var editArg db.EditCartItemByIDParams
	editArg.ID = cartItem.ID
	if cartItem.Quantity == 20 {
		Err = append(Err, "cart item already added with maximum possible quantity. Not changing the quantity")
		editArg.Quantity = cartItem.Quantity
	} else if cartItem.Quantity < product.Stock {
		editArg.Quantity = cartItem.Quantity + 1
	} else {
		editArg.Quantity = product.Stock
	}
	editItem, err := u.DB.EditCartItemByID(context.TODO(), editArg)
	if err != nil {
		log.Warn("internal error editing cartItem on adding cartItem on already existing item", err.Error())
		http.Error(w, "internal error updating cartItem on adding the cartItem again.", http.StatusInternalServerError)
		return
	}

	var respEditItem respCartItem
	respEditItem.CartID = cartItem.ID
	respEditItem.ProductID = product.ID
	respEditItem.ProductName = product.Name
	respEditItem.Quantity = int(editItem.Quantity)
	var resp struct {
		Data    respCartItem `json:"data"`
		Message string       `json:"message"`
		Err     []string     `json:"errors"`
	}
	resp.Data = respEditItem
	resp.Message = "successfully updated cart item on adding the cart item on already existing cart item"
	resp.Err = Err
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) EditCartHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	// get the reques body and validate the body and it's fields
	var req struct {
		ProductID uuid.UUID `json:"product_id"`
		Quantity  int       `json:"quantity"`
	}
	type respCartItem struct {
		CartID      uuid.UUID `json:"cart_id"`
		ProductID   uuid.UUID `json:"product_id"`
		ProductName string    `json:"product_name"`
		Quantity    int       `json:"quantity"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request data format", http.StatusBadRequest)
		return
	} else if req.Quantity < 0 {
		http.Error(w, "invalid quantity", http.StatusBadRequest)
		return
	} else if req.Quantity > 20 {
		http.Error(w, "invalid quantity. maximum possible quantity is 20", http.StatusBadRequest)
		return
	}

	// create Err slice to give as resposne errors
	var Err []string
	// check if the productID updating in cart is of a valid product
	product, err := u.DB.GetProductByID(context.TODO(), req.ProductID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid productID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("internal error fetching product from cartID:", err.Error())
		http.Error(w, "internal server error fetching product from cartID", http.StatusInternalServerError)
		return
	}

	// get cartItem for the product and check if it already exists or not
	// if so update the cartItem
	// else add the cartItem with the productID and quantity
	var getArg db.GetCartItemByUserIDAndProductIDParams
	getArg.ProductID = product.ID
	getArg.UserID = user.ID
	cartItem, err := u.DB.GetCartItemByUserIDAndProductID(context.TODO(), getArg)
	// add product if the product is not in carts;
	if err == sql.ErrNoRows {
		if req.Quantity == 0 {
			http.Error(w, "invalid quantity to add a product on editHandler. change quantity > 0", http.StatusBadRequest)
			return
		}
		var arg db.AddCartItemParams
		arg.UserID = user.ID
		arg.ProductID = req.ProductID
		if product.Stock == 0 {
			http.Error(w, "trying to edit and add an out of stock product", http.StatusBadRequest)
			return
		} else if req.Quantity > int(product.Stock) {
			Err = append(Err, "adding more quantity of product than there is stock. Reverting the quantity back to maximum allotable")
			arg.Quantity = product.Stock
		} else {
			arg.Quantity = int32(req.Quantity)
		}
		arg.ProductID = product.ID
		arg.UserID = user.ID

		// add the product to the cart
		item, err := u.DB.AddCartItem(context.TODO(), arg)
		if err != nil {
			log.Warn("internal error adding cartItem:", err.Error())
			http.Error(w, "internal error adding cartItem", http.StatusInternalServerError)
			return
		}

		// give back response and handle error cases
		var respItem = respCartItem{
			CartID:      item.ID,
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    int(item.Quantity),
		}
		var resp struct {
			Data    respCartItem `json:"data"`
			Message string       `json:"message"`
			Err     []string     `json:"errors"`
		}
		resp.Data = respItem
		resp.Message = "successfully added cart item since already didn't exist."
		resp.Err = Err

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return

		// handle the internal error case on fetching the cartItem
		// to check if the productItem is already in the cart or not
	} else if err != nil {
		log.Warn("error fetching cartItem before editing cartItem using GetCartItemByID:", err.Error())
		http.Error(w, "internal error fetching cart item before editing", http.StatusInternalServerError)
		return
	}

	// edit cart item when there is an existing cart item for the product
	var editArg db.EditCartItemByIDParams
	editArg.ID = cartItem.ID
	if req.Quantity == 0 {
		var deleteCartArg db.DeleteCartItemByUserIDAndProductIDParams
		deleteCartArg.ProductID = product.ID
		deleteCartArg.UserID = user.ID
		err = u.DB.DeleteCartItemByUserIDAndProductID(context.TODO(), deleteCartArg)
		if err != nil {
			log.Warn("error deleting item from carItem when quantity == 0 in EditCartHandler:", err.Error())
			http.Error(w, "internal error deleting cartItem when qunatity is made zero", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		msg := "successfully deleted cartItem on zero quantity"
		w.Write([]byte(msg))
		return
	} else if req.Quantity > int(product.Stock) {
		Err = append(Err, "edit cartItem with more quantity than there is stock. Reallocation the cartItem to the maximum possible")
		editArg.Quantity = product.Stock
	} else {
		editArg.Quantity = int32(req.Quantity)
	}
	// edit the cartItem
	editedItem, err := u.DB.EditCartItemByID(context.TODO(), editArg)
	if err == sql.ErrNoRows {
		log.Warn("no rows edited after successful query in cart at EditCartHandler:", err.Error())
		http.Error(w, "internal error editing cartItem", http.StatusInternalServerError)
		return
	} else if err != nil {
		log.Warn("internal error editing cartItem: ", editArg, err.Error())
		http.Error(w, "internal error editing cartItem", http.StatusInternalServerError)
		return
	}

	// give back response on successful editing of cartItem
	var respItem respCartItem
	respItem.CartID = cartItem.ID
	respItem.ProductID = product.ID
	respItem.ProductName = product.Name
	respItem.Quantity = int(editedItem.Quantity)
	var resp struct {
		Data    respCartItem `json:"data"`
		Message string       `json:"message"`
		Err     []string     `json:"errors"`
	}
	resp.Data = respItem
	resp.Message = "successfully edited cartItem"
	resp.Err = Err
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) DeleteCartHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	// get the request body and validate the request body and it's fields
	var req struct {
		ProductID uuid.UUID `json:"product_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request data", http.StatusBadRequest)
		return
	}
	// fetch product to check if it's a valid product
	product, err := u.DB.GetProductByID(context.TODO(), req.ProductID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid productID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("internal erro fetching product to delete it from cart:", err.Error())
		http.Error(w, "internal error fetching product to delete it from cart", http.StatusInternalServerError)
		return
	}

	// get the cartItem for the product to check if the product
	// exists in the user cart
	var getItemArg db.GetCartItemByUserIDAndProductIDParams
	getItemArg.ProductID = product.ID
	getItemArg.UserID = user.ID
	_, err = u.DB.GetCartItemByUserIDAndProductID(context.TODO(), getItemArg)
	if err == sql.ErrNoRows {
		http.Error(w, "trying to delete non-existent product from cart", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "internal server error fetching cartItem to check if the product exists in cart", http.StatusInternalServerError)
		return
	}

	// make deleteArg to deleteCartItem
	// deletes the cartItem if there is a matching product
	var deleteArg db.DeleteCartItemByUserIDAndProductIDParams
	deleteArg.ProductID = product.ID
	deleteArg.UserID = user.ID
	err = u.DB.DeleteCartItemByUserIDAndProductID(context.TODO(), deleteArg)
	if err != nil {
		log.Warn("internal error deleting cartItem with valid productID:", err.Error())
		http.Error(w, "internal error deleting cartItem", http.StatusInternalServerError)
		return
	}

	// send the response on successful deletion of product from cart
	w.Header().Set("Content-Type", "text/plain")
	message := fmt.Sprintf("product: %s deleted successfully from cartItems", product.Name)
	w.Write([]byte(message))
}

// //////////////////////////
// order handlers
func (u *User) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	// get orders by userId
	orders, err := u.DB.GetOrdersByUserID(context.TODO(), user.ID)
	if err != nil {
		log.Warn("error fetching orders by userID in getOrdersHandler:", err.Error())
		http.Error(w, "internal error fetching orders", http.StatusInternalServerError)
		return
	}

	// respOrderItem struct
	type respOrderItem struct {
		OrderItemID uuid.UUID `json:"order_item_id"`
		Status      string    `json:"order_status"`
		ProductID   uuid.UUID `json:"product_id"`
		ProductName string    `json:"product_name"`
		Price       float64   `json:"price"`
		Quantity    int       `json:"quantity"`
		TotalAmount float64   `json:"total_amount"`
	}
	// respOrder struct
	type respOrder struct {
		OrderID       uuid.UUID       `json:"order_id"`
		PaymentMethod string          `json:"payment_method"`
		PaymentStatus string          `json:"payment_status"`
		OrderItems    []respOrderItem `json:"order_items"`
	}

	// var respOrders, errors
	var respOrders []respOrder
	var Err []string

	for _, o := range orders {
		var temp respOrder
		orderItems, err := u.DB.GetOrderItemsByOrderID(context.TODO(), o.ID)
		if err != nil {
			log.Warn("error fetching orderItem in GetOrderHandler:", err.Error())
			Err = append(Err, "error fetching order by orderID:", o.ID.String())
		} else {
			temp.OrderID = o.ID
			payment, err := u.DB.GetPaymentByOrderID(context.TODO(), o.ID)
			if err != nil {
				log.Warn("error fetching payment for orderID in GetOrdersHandler:" + err.Error())
			}
			temp.PaymentMethod = payment.Method
			temp.PaymentStatus = payment.Status
			for _, oi := range orderItems {
				var orderItem respOrderItem
				orderItem.OrderItemID = oi.ID
				orderItem.Status = oi.Status
				orderItem.ProductID = oi.ProductID
				orderItem.ProductName = oi.ProductName
				orderItem.Price = oi.Price
				orderItem.Quantity = int(oi.Quantity)
				orderItem.TotalAmount = oi.TotalAmount
				temp.OrderItems = append(temp.OrderItems, orderItem)
			}
			respOrders = append(respOrders, temp)
		}
	}

	// send response
	var resp struct {
		Data    []respOrder `json:"data"`
		Err     []string    `json:"errors"`
		Message string      `json:"message"`
	}
	resp.Data = respOrders
	resp.Err = Err
	resp.Message = "successfully fetched orders and orderItems"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) GetOrderItemsHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	orderItems, err := u.DB.GetOrderItemsByUserID(context.TODO(), user.ID)
	if err != nil {
		log.Warn("error fetching order_items from userID in getOrderItems:", err.Error())
		http.Error(w, "internal error fetching orderItems", http.StatusInternalServerError)
		return
	}
	type respOrderItem struct {
		OrderItemID uuid.UUID `json:"order_item_id"`
		Status      string    `json:"status"`
		ProductID   uuid.UUID `json:"product_id"`
		Price       float64   `json:"price"`
		Quantity    int       `json:"quantity"`
		TotalAmount float64   `json:"total_amount"`
	}
	// storing response Order items
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
		OrderItems []respOrderItem `json:"order_items"`
		Messaege   string          `json:"message"`
	}
	resp.OrderItems = respOrderItems
	resp.Messaege = "successfully fetched orderItems"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) AddCartToOrderHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	// check if the user has a phone number
	// return error if user has no registered phone number
	if !user.Phone.Valid {
		http.Error(w, "phone number not added for user. Unauthorized to make an order", http.StatusBadRequest)
		return
	}

	// get the shipping address from the request
	ShippingAddressIDStr := r.URL.Query().Get("shipping_address_id")
	ShippingAddressID, err := uuid.Parse(ShippingAddressIDStr)
	if err != nil {
		http.Error(w, "invalid address format", http.StatusBadRequest)
		return
	}

	// to add and display insignificant errors
	var Err []string
	var Messages []string

	// get coupon
	var ifCouponExists bool
	couponName := r.URL.Query().Get("coupon_name")
	coupon, err := u.DB.GetCouponByName(context.TODO(), couponName)
	if couponName == "" {
		// simply put the if condition to not the ErrNoRows error when the couponName is empty
	} else if err == sql.ErrNoRows {
		http.Error(w, "invalid coupon applied. Either leave that empty or apply valid coupons", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Error("error fetching coupon from couponName in AddCartToOrderHandler:", err.Error())
		http.Error(w, "internal error fetching coupon", http.StatusInternalServerError)
		return
	} else {
		ifCouponExists = true
	}

	// get a valid address for the shipping address
	address, err := u.DB.GetAddressByID(context.TODO(), ShippingAddressID)
	if err == sql.ErrNoRows {
		http.Error(w, "not a valid addressID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error fetching address by id for order:", err.Error())
	} else if address.UserID != user.ID {
		http.Error(w, "shipping_addresss_id is not a valid id for the user", http.StatusBadRequest)
		return
	}

	// for future calculations
	var discountAmount float64
	var ifCouponValid bool

	totalAmount, err := u.DB.GetSumOfCartItemsByUserID(context.TODO(), user.ID)
	if err != nil {
		log.Error("error fetching totalAmount for user in AddCartToOrderHandler:", err.Error())
		http.Error(w, "internal error fetching necessary data", http.StatusInternalServerError)
		return
	}
	fmt.Println(discountAmount, ifCouponExists, ifCouponValid, coupon)

	// set value for discountAmount
	// also set ifCouponValid true if coupon exists
	if ifCouponExists && coupon.TriggerPrice <= totalAmount &&
		time.Now().After(coupon.StartDate) && time.Now().Before(coupon.EndDate) && !coupon.IsDeleted {
		fmt.Println("entered here")
		ifCouponValid = true
		if coupon.DiscountType == utils.CouponDiscountTypeFlat {
			discountAmount = coupon.DiscountAmount
		} else if coupon.DiscountType == utils.CouponDiscountTypePercentage {
			discountAmount = coupon.DiscountAmount * totalAmount / 100
		}
	}

	fmt.Println(discountAmount, ifCouponExists, ifCouponValid, coupon)
	// take the payment method from url query
	paymentMethod := r.URL.Query().Get("payment_method")
	if paymentMethod == utils.StatusPaymentMethodWallet {
		wallet, err := u.DB.GetWalletByUserID(context.TODO(), user.ID)
		if err == sql.ErrNoRows {
			log.Error("error no wallet exists for user in AddCartToOrderHandler for user:", err.Error())
			http.Error(w, "internal error. Wallet is yet to be provided for the user", http.StatusInternalServerError)
			return
		} else if err != nil {
			log.Error("error fetching wallet for user in AddCartToOrderHandler for user:", err.Error())
			http.Error(w, "internal error: failed to fetch wallet for user", http.StatusInternalServerError)
			return
		}

		// make sure coupon is in active time and it is valid for the user
		// change the values of discountAmount, ifCouponValid, totalAmount
		// accordingly
		if wallet.Savings < totalAmount {
			msg := fmt.Sprintf("not enough money in wallet to buy product \n"+
				"Needed: %0.2f; Your wallet has %0.2f", totalAmount, wallet.Savings)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

	}

	// get the cart items
	cartItems, err := u.DB.GetCartItemsByUserID(context.TODO(), user.ID)
	if err != nil {
		log.Warn("error fetching cartItems to add to order_items:", err.Error())
		http.Error(w, "internal error fetching cartItems to add to order", http.StatusInternalServerError)
		return
	} else if len(cartItems) == 0 {
		http.Error(w, "no cart items. Cannot place an order", http.StatusBadRequest)
		return
	}

	// add an order
	order, err := u.DB.AddOrder(context.TODO(), user.ID)
	if err != nil {
		log.Warn("error creating order:", err.Error())
		http.Error(w, "internal error creating order", http.StatusInternalServerError)
		return
	}

	// add shipping address for order
	var addrArg db.AddShippingAddressParams
	addrArg.OrderID = order.ID
	addrArg.HouseName = address.BuildingName
	addrArg.StreetName = address.StreetName
	addrArg.Town = address.Town
	addrArg.District = address.District
	addrArg.State = address.State
	addrArg.Pincode = address.Pincode
	shipAddr, err := u.DB.AddShippingAddress(context.TODO(), addrArg)
	if err != nil {
		log.Warn("error adding shipping address for order. deleting order:", err.Error())
		dErr := u.DB.DeleteOrderByID(context.TODO(), order.ID)
		if dErr != nil {
			log.Warn("error deleting order after failing to add shipping address for order:", err.Error())
		}
		http.Error(w, "internal error adding shipping address for the order", http.StatusInternalServerError)
		return
	}

	// add cartItems to orderItems
	for _, v := range cartItems {
		var addArg db.AddOrderITemParams
		addArg.OrderID = order.ID
		addArg.ProductID = v.ProductID
		addArg.Price = v.Price
		addArg.Quantity = v.Quantity
		orderItem, err := u.DB.AddOrderITem(context.TODO(), addArg)
		if err != nil {
			log.Warn("error adding cartItem to order_item:", err.Error())
			http.Error(w, "internal error adding cartItem to order_items", http.StatusInternalServerError)
			return
		}
		product, err := u.DB.GetProductByID(context.TODO(), v.ProductID)
		if err == sql.ErrNoRows {
			log.Warn("no product with matching productId from cartItem:", err.Error())
		} else if err != nil {
			log.Warn("error executing GetProductByID query:", err.Error())
		}

		// get sellerID from OrderItemID
		sellerID, err := u.DB.GetSellerIDFromOrderItemID(context.TODO(), orderItem.ID)
		if err != nil {
			log.Warn("error fetching sellerID from orderItemID:", err.Error())
		}
		// add vendor payment for each orderItem
		var addVendorPayArg db.AddVendorPaymentParams
		addVendorPayArg.OrderItemID = orderItem.ID
		addVendorPayArg.SellerID = sellerID
		addVendorPayArg.Status = utils.StatusVendorPaymentWaiting
		addVendorPayArg.TotalAmount = orderItem.TotalAmount
		addVendorPayArg.PlatformFee = orderItem.TotalAmount * utils.PlatformFeePercentage
		addVendorPayArg.CreditAmount = orderItem.TotalAmount * (1 - utils.PlatformFeePercentage)
		_, err = u.DB.AddVendorPayment(context.TODO(), addVendorPayArg)
		if err == sql.ErrNoRows {
			log.Warn("error no vendorPayment done after successful query:", err.Error())
		} else if err != nil {
			log.Warn("error failed addVendorPayment in AddCartToOrderHandler:", err.Error())
		}
		var decArg db.DecProductStockByIDParams
		decArg.ProductID = v.ProductID
		decArg.DecQuantity = v.Quantity
		decProduct, err := u.DB.DecProductStockByID(context.TODO(), decArg)
		if err != nil {
			log.Warn("error decrementing from product stock after placing order:", err.Error())
		} else {
			msg := fmt.Sprintf("decremented product: %s from quantity: %d to %d.", v.ProductID.String(), product.Stock, decProduct.Stock)
			log.Info(msg)
		}
	}

	// update order total and discount amount
	var editOrderAmountArg db.EditOrderAmountByIDParams
	editOrderAmountArg.ID = order.ID
	editOrderAmountArg.TotalAmount = totalAmount
	editOrderAmountArg.DiscountAmount = discountAmount
	if ifCouponValid {
		editOrderAmountArg.CouponID.Valid = true
		editOrderAmountArg.CouponID.UUID = coupon.ID
	}
	updatedOrder, err := u.DB.EditOrderAmountByID(context.TODO(), editOrderAmountArg)
	if err != nil {
		log.Error("error updating order Total amount:", err.Error())
	}
	// deleting cartItems after successfully adding cartItems to order
	err = u.DB.DeleteCartItemsByUserID(context.TODO(), user.ID)
	if err != nil {
		message := "error deleting the cart items after successfully adding it to the order_items:"
		log.Warn(message, err.Error())
		Err = append(Err, message)
	}

	// add sumTotal to payments for the order_id
	var payArg db.AddPaymentParams

	if paymentMethod == utils.StatusPaymentMethodRpay {
		payArg.Method = utils.StatusPaymentMethodRpay
	} else if paymentMethod == utils.StatusPaymentMethodWallet {
		payArg.Method = utils.StatusPaymentMethodWallet
	} else {
		paymentMethod = utils.StatusPaymentMethodCod
	}

	payArg.OrderID = order.ID
	payArg.Status = utils.StatusPaymentProcessing
	payArg.TotalAmount = updatedOrder.NetAmount
	payment, err := u.DB.AddPayment(context.TODO(), payArg)
	if err != nil {
		log.Warn("error adding payment for the order")
		http.Error(w, "error adding payment for the order", http.StatusInternalServerError)
		return
	}

	orderItems, err := u.DB.GetOrderItemsByOrderID(context.TODO(), order.ID)
	if err != nil {
		message := "error fetching orderItems after successfully adding orderItems:"
		log.Warn(message, err.Error())
		Err = append(Err, message)
	}

	// update wallet total savings if the payment is done through wallet
	if paymentMethod == utils.StatusPaymentMethodWallet {
		var retractArg db.RetractSavingsFromWalletByUserIDParams
		retractArg.Savings = order.NetAmount
		retractArg.UserID = user.ID
		updatedWallet, err := u.DB.RetractSavingsFromWalletByUserID(context.TODO(), retractArg)
		if err != nil {
			log.Error(
				"error retracting savings from wallet after placing order"+
					"via wallet in AddCartToOrderHandler:", err.Error())
		} else {
			msg := fmt.Sprintf("retracted %0.2f from wallet;\n", order.NetAmount) +
				fmt.Sprintf("Wallet balance: %0.2f ", updatedWallet.Savings)
			Messages = append(Messages, msg)
		}
	}

	Messages = append(Messages, "successfully added the cart items to orders")
	var resp struct {
		Order           db.Order                       `json:"order"`
		Phone           int                            `json:"phone"`
		Payment         db.Payment                     `json:"payment"`
		OrderItems      []db.GetOrderItemsByOrderIDRow `json:"order_items"`
		ShippingAddress db.AddShippingAddressRow       `json:"shipping_address"`
		Err             []string                       `json:"error"`
		Messages        []string                       `json:"messages"`
	}

	resp.Phone = int(user.Phone.Int64)
	resp.Order = updatedOrder
	resp.Payment = payment
	resp.OrderItems = orderItems
	resp.ShippingAddress = shipAddr
	resp.Err = Err
	resp.Messages = Messages
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) CancelOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	// get request ordreItemID from request
	orderItemIDStr := r.URL.Query().Get("order_item_id")
	orderItemID, err := uuid.Parse(orderItemIDStr)
	if err != nil {
		http.Error(w, "ordreItemID not in uuid format", http.StatusBadRequest)
		return
	}
	// get OrderItemId
	orderItem, err := u.DB.GetOrderItemByID(context.TODO(), orderItemID)
	if err == sql.ErrNoRows {
		http.Error(w, "not a valid order_item_id", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error fetching orderItemByID in cancel order:", err.Error())
		http.Error(w, "internal server error fetching orderItem", http.StatusInternalServerError)
		return
	}

	// get userID from orderItemID and verify it's the same user's orderItem
	OrderItemUserID, err := u.DB.GetUserIDFromOrderItemID(context.TODO(), orderItemID)
	if err != nil {
		log.Warn("error fetching userID from orderItemID in cancel order:", err.Error())
		http.Error(w, "internal error fetching userID from orderItemID", http.StatusInternalServerError)
		return
	} else if OrderItemUserID != user.ID {
		http.Error(w, "not user's orderItemID. User unauthorized.", http.StatusUnauthorized)
		return
	}
	// check if coupon is applied to the given orderItem order
	// if applied deny request
	order, err := u.DB.GetOrderByID(context.TODO(), orderItem.OrderID)
	if err != nil {
		log.Error("error fetching order by orderItem in CancelOrderItemHandler for user:", err.Error())
	} else if order.CouponID.Valid {
		http.Error(w, "coupon applied; cannot cancel a single item from order", http.StatusBadRequest)
		return
	}

	// make Err, Messages slice to give for response errors, messages
	var Err []string
	var Messages []string

	// check if the order is shipped or not
	// if order is processing or pending, cancel order
	if orderItem.Status == utils.StatusOrderCancelled {
		http.Error(w, "order already cancelled. cannot cancel", http.StatusForbidden)
		return
	} else if orderItem.Status == utils.StatusOrderDelivered {
		http.Error(w, "order already delivered. cannot cancel", http.StatusForbidden)
		return
	} else if orderItem.Status == utils.StatusOrderShipped {
		http.Error(w, "order already shipped. cannot cancel", http.StatusForbidden)
		return
	} else if orderItem.Status == utils.StatusOrderPending || orderItem.Status == utils.StatusOrderProcessing {
		// cancel order
		var editOrderItemArg db.EditOrderItemStatusByIDParams
		editOrderItemArg.ID = orderItem.ID
		editOrderItemArg.Status = utils.StatusOrderCancelled
		_, err = u.DB.EditOrderItemStatusByID(context.TODO(), editOrderItemArg)
		if err != nil {
			log.Warn("error editing orderItemStatus:", err.Error())
			http.Error(w, "internal error editing orderItemStatus", http.StatusInternalServerError)
			return
		}
		// add money to wallet in case it is already paid
		payment, err := u.DB.GetPaymentByOrderID(context.TODO(), order.ID)
		if err != nil {
			log.Error("error fetching payment from orderID in CancelOrderItemHandler:", err.Error())
			http.Error(w, "internal error fetching necessary items to cancel order", http.StatusInternalServerError)
			return
		}
		if payment.Status == utils.StatusPaymentSuccessful {
			wallet, err := u.DB.AddSavingsToWalletByUserID(context.TODO(), db.AddSavingsToWalletByUserIDParams{
				Savings: orderItem.TotalAmount,
				UserID:  user.ID,
			})
			if err != nil {
				log.Error("error updating money back to wallet on cancelling orderItem:", err.Error())
				Err = append(Err, "error adding money to wallet after cancelling order")
			} else {
				msg := fmt.Sprintf("successfully added amount: %0.2f back to wallet.\nCurrent balance: %0.2f",
					orderItem.TotalAmount, wallet.Savings)
				Messages = append(Messages, msg)
			}

		}
		// increment product stock after cancelling order
		var incArg db.IncProductStockByIDParams
		incArg.ProductID = orderItem.ProductID
		incArg.IncQuantity = orderItem.Quantity
		_, err = u.DB.IncProductStockByID(context.TODO(), incArg)
		if err != nil {
			log.Warn("error incrementing product stock after cancelling order:", err.Error())

			// print the increment quantity if there were no errors incrementing
		} else {
			product, err := u.DB.GetProductByID(context.TODO(), orderItem.ProductID)
			if err != nil {
				msg := fmt.Sprintf("incremented product stock after cancelling: %d added back", incArg.IncQuantity)
				log.Info(msg)
			} else {
				msg := fmt.Sprintf("incremented product stock after cancelling: from stock: %d to new_stock: %d", product.Stock-incArg.IncQuantity, product.Stock)
				log.Info(msg)
			}
		}

		// decrement payment on cancelling order
		payment, err = u.DB.DecPaymentAmountByOrderItemID(context.TODO(), orderItem.ID)
		if err != nil {
			log.Warn("error updating payment for the order:", err.Error())
			Err = append(Err, "error updating payment for the order after cancelling item")

			// change the payment status to returned if the payment becomes zero
		} else if payment.TotalAmount == 0 {
			var editPaymentArg db.EditPaymentStatusByIDParams
			editPaymentArg.ID = payment.ID
			editPaymentArg.Status = utils.StatusPaymentReturned
			zeroPayment, err := u.DB.EditPaymentStatusByID(context.TODO(), editPaymentArg)
			if err != nil {
				log.Warn("error editing payment status:", err.Error())
			} else {
				payment.Status = zeroPayment.Status
			}
		}
		err = u.DB.CancelVendorPaymentByOrderItemID(context.TODO(), orderItem.ID)
		if err != nil {
			log.Error("error cancelling vendor payment in CancelOrderItem for user:", err.Error())
		}

		type RespPayment struct {
			PaymentID      uuid.UUID `json:"payment_id"`
			NewTotalAmount float64   `json:"new_total_amount"`
			PaymentMethod  string    `json:"payment_method"`
			PaymentStatus  string    `json:"payment_status"`
		}
		var rPay RespPayment
		rPay.PaymentID = payment.ID
		rPay.NewTotalAmount = payment.TotalAmount
		rPay.PaymentStatus = payment.Status
		rPay.PaymentMethod = payment.Method

		// send response after successful order cancellation
		var resp struct {
			OrderItemID uuid.UUID   `json:"order_item_id"`
			Messages    []string    `json:"messages"`
			Err         []string    `json:"errors"`
			NewPayment  RespPayment `json:"new_payment"`
		}
		resp.OrderItemID = orderItem.ID
		resp.Messages = append(Messages, "order_item has been cancelled.")
		resp.Err = Err
		resp.NewPayment = rPay
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

		// incase the orderItem status doesn't match all the rest// theoretically impossible
	} else {
		log.Warn("invalid order status for orderItem in cancel order:")
		http.Error(w, "invalid order status for orderItem", http.StatusInternalServerError)
		return
	}
}

func (u *User) CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	var messages []string
	var errors []string

	// get orderID from request params
	orderIDStr := r.URL.Query().Get("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		http.Error(w, "orderID not in uuid format", http.StatusBadRequest)
		return
	}

	// check whether it is a valid order and if it the current users's order
	order, err := u.DB.GetOrderByID(context.TODO(), orderID)
	if err == sql.ErrNoRows {
		http.Error(w, "not a valid orderID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error fetching order by ID in CancelOrderHandler:", err.Error())
		http.Error(w, "error fetching order by id", http.StatusInternalServerError)
		return
	} else if order.UserID != user.ID {
		http.Error(w, "not user's order to cancel", http.StatusUnauthorized)
		return
	}

	payment, err := u.DB.GetPaymentByOrderID(context.TODO(), order.ID)
	if err != nil {
		log.Warn("error fetching payment from orderID in CancelOrderHandler")
	} else if payment.Status == utils.StatusPaymentSuccessful {
		// add money back to wallet on already paid order
		// and cancel that payment
		var arg db.AddSavingsToWalletByUserIDParams
		arg.Savings = payment.TotalAmount
		arg.UserID = user.ID
		_, err = u.DB.AddSavingsToWalletByUserID(context.TODO(), arg)
		if err != nil {
			log.Warn("error transferring refund amount to wallet:", err.Error())
			messages = append(messages, "error transferring refund amount to wallet")
		} else {
			log.Warn("successfully transferred amount to user wallet")
			messages = append(messages, fmt.Sprintf("successfully transferred cancellation refund amount: %.2f to wallet", payment.TotalAmount))
			// cancel payment on successful addition of money to wallet
			_, err = u.DB.CancelPaymentByOrderID(context.TODO(), orderID)
			if err != nil {
				log.Warn("error returning payment by orderID in CancelOrderHandler:", err.Error())
				http.Error(w, "intenral error cancelling payment by orderID after successfully cancelling orderItems", http.StatusInternalServerError)
				return
			} else {
				msg := "successfully cancelled payment for the cancelled order"
				messages = append(messages, msg)
			}
		}
	}
	orderItems, err := u.DB.GetOrderItemsByOrderID(context.TODO(), order.ID)
	if err != nil {
		log.Warn("error fetching orderItems by ordreID")
	} else {
		for _, oi := range orderItems {
			var vendorPayArg db.EditVendorPaymentStatusByOrderItemIDParams
			vendorPayArg.OrderItemID = oi.ID
			vendorPayArg.Status = utils.StatusVendorPaymentCancelled
			_, err = u.DB.EditVendorPaymentStatusByOrderItemID(context.TODO(), vendorPayArg)
			if err != nil {
				log.Warn("error cancelling vendor payment for orderItem:", order.ID.String())
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	var resp struct {
		Messages []string `json:"messages"`
		Errors   []string `json:"errors"`
	}
	resp.Messages = append(messages, "successfully cancelled order")
	resp.Errors = errors
	json.NewEncoder(w).Encode(resp)
}

func (u *User) MakeOnlinePaymentHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	orderIDStr := r.URL.Query().Get("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		http.Error(w, "invalid orderID", http.StatusInternalServerError)
		return
	}
	order, err := u.DB.GetOrderByID(context.TODO(), orderID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid orderID", http.StatusBadRequest)
		return
	}

	// RazorpayData holds the dynamic values to be injected into the HTML template
	type RazorpayData struct {
		Key           string
		Amount        int
		Currency      string
		EcomName      string
		Description   string
		OrderID       string    // for storing orderID we get from razorpay
		DBOrderID     uuid.UUID // to store our orderID
		Username      string
		Email         string
		Contact       string
		Errors        []string
		DisplayAmount float64
	}

	payment, err := u.DB.GetPaymentByOrderID(context.TODO(), order.ID)
	var errors []string
	if err != nil {
		http.Error(w, "internal error fetching payment for order", http.StatusInternalServerError)
		log.Warn("error fetching payment for valid orderID in MakeOnlinePaymentHandler:", err.Error())
		return
	} else if payment.Status == utils.StatusPaymentCancelled {
		http.Error(w, "order already cancelled", http.StatusBadRequest)
		return
	} else if payment.Status == utils.StatusPaymentSuccessful {
		http.Error(w, "order payment already successful", http.StatusBadRequest)
		return
	} else if payment.CreatedAt.Before(time.Now().Add(-10 * time.Minute)) {
		var editPaymentStatusArg db.EditPaymentStatusByIDParams
		editPaymentStatusArg.ID = payment.ID
		editPaymentStatusArg.Status = utils.StatusPaymentCancelled
		_, err := u.DB.EditPaymentStatusByID(context.TODO(), editPaymentStatusArg)
		if err != nil {
			errors = append(errors, "internal error updating payment status to cancelled")
			log.Warn("error updating payment status to cancelled in MakeOnlinePayment Handler")
		}
		_, err = u.DB.CancelOrderByID(context.TODO(), orderID)
		if err != nil {
			errors = append(errors, "error cancelling order_items for invalid orderID without timeout payment error")
			log.Warn("internal error cancelling order items for order placed before 10 minutes for razorpay payment.")
		}

		orderItems, err := u.DB.GetOrderItemsByOrderID(context.TODO(), orderID)
		if err != nil {
			errors = append(errors, "error fetching order_items for invalid orderID without timeout payment error")
			log.Warn("internal error cancelling order items for order placed before 10 minutes for razorpay payment.")
		}

		for _, oi := range orderItems {
			var arg db.IncProductStockByIDParams
			arg.IncQuantity = oi.Quantity
			arg.ProductID = oi.ProductID
			_, err = u.DB.IncProductStockByID(context.TODO(), arg)
			if err != nil {
				log.Error("error incrementing stock for cancelled order_item in MakeOnlinePaymentHandler")
			} else {
				log.Info("successfully restocked product " + oi.ProductName + " on cancel order:")
			}

			// cancel vendor payments for the respective orders
			var vendorPayArg db.EditVendorPaymentStatusByOrderItemIDParams
			vendorPayArg.OrderItemID = oi.ID
			vendorPayArg.Status = utils.StatusVendorPaymentCancelled
			_, err = u.DB.EditVendorPaymentStatusByOrderItemID(context.TODO(), vendorPayArg)
			if err != nil {
				log.Warn("error cancelling vendor payment for orderItem:", order.ID.String())
			}
		}
		http.Error(w, "cancelled order and payment since time limit exceeded!"+strings.Join(errors, "\n"), http.StatusBadRequest)
		return
	}

	// fetch razorpay to give unique orderID that is to be used with its api
	// in the rpay template when clicking the pay with razorpay button
	envM, err := env.Read(".env")
	if err != nil {
		log.Fatal("error loading .env file in MakeOnlinePaymentHandler")
	}
	rpOrderIDStr, err := helpers.ExecuteRazorpay(payment.TotalAmount)
	if err != nil {
		log.Warn("error executing razorpay")
		http.Error(w, "internal error executing razorpay", http.StatusInternalServerError)
		return
	}

	// data for the razorpaytemplate
	data := RazorpayData{
		Key:           envM[envname.RPID],
		Amount:        int(payment.TotalAmount) * 100, // Amount in paise (₹500)
		Currency:      "INR",
		EcomName:      utils.EcomName,
		Description:   "Purchase of toys",
		OrderID:       rpOrderIDStr,
		DBOrderID:     order.ID,
		Username:      user.Name,
		Email:         user.Email,
		Contact:       strconv.Itoa(int(user.Phone.Int64)),
		Errors:        errors,
		DisplayAmount: payment.TotalAmount,
	}

	// Parse the template file
	tmpl, err := template.ParseFiles("./static/template/rpay.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Println("Template parsing error:", err)
		return
	}

	// Execute the template and pass data
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Println("Template execution error:", err)
	}

}

func (u *User) PaymentSuccessHandler(w http.ResponseWriter, r *http.Request) {
	type RazorpayResponse struct {
		PaymentID    string `json:"payment_id"`
		OrderID      string `json:"order_id"`
		Signature    string `json:"signature"`
		DBOrderIDStr string `json:"db_order_id"`
	}
	var resp RazorpayResponse
	json.NewDecoder(r.Body).Decode(&resp)

	if helpers.VerifyRazorpaySignature(resp.OrderID, resp.PaymentID, resp.Signature) {
		DBOrderID, err := uuid.Parse(resp.DBOrderIDStr)
		if err != nil {
			log.Warn("error fetching the dbOrderID from the razorpayResponse in PaymentSuccessHandler")
			http.Error(w, "error fetching dbOrderID from razorPay to update payment success", http.StatusInternalServerError)
			return
		}
		var editPaymentArg db.EditPaymentByOrderIDParams
		editPaymentArg.OrderID = DBOrderID
		editPaymentArg.Status = utils.StatusPaymentSuccessful
		editPaymentArg.TransactionID.String = resp.PaymentID
		editPaymentArg.TransactionID.Valid = true
		payment, err := u.DB.EditPaymentByOrderID(context.TODO(), editPaymentArg)
		if err != nil {
			log.Warn("error updating the payment after successful payment using razorpay")
			http.Error(w, "internal error updating payment after successful payment using razorpay", http.StatusInternalServerError)
			return
		}
		msg := "successfully updated payment status for the order." +
			"payment method: " + payment.Method + "\n" +
			"order id: " + payment.OrderID.String()
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte(msg))
		return
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		msg := "failed to verify payment"
		w.Write([]byte(msg))
		log.Info("Payment verification failed:", resp.PaymentID)
		return
	}
}

func (u *User) ReturnOrderHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	orderIdStr := r.URL.Query().Get("order_id")
	orderID, err := uuid.Parse(orderIdStr)
	if err != nil {
		http.Error(w, "wrong orderID format", http.StatusBadRequest)
		return
	}
	order, err := u.DB.GetOrderByID(context.TODO(), orderID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid orderID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Error("error fetching order by orderID")
		http.Error(w, "internal error fetching order by orderID", http.StatusInternalServerError)
		return
	} else if order.UserID != user.ID {
		http.Error(w, "not the current users's order. Unauthorized", http.StatusUnauthorized)
		return
	}
	payment, err := u.DB.GetPaymentByOrderID(context.TODO(), order.ID)
	if err != nil {
		log.Warn("error fetching payemnt from ordreID in return ordr handler")
		http.Error(w, "internal error fetching payment from orderID", http.StatusInternalServerError)
		return
	} else if payment.Status == utils.StatusPaymentReturned {
		http.Error(w, "items already returned and user is refunded", http.StatusBadRequest)
		return
	} else if payment.Status == utils.StatusPaymentSuccessful {
		// exit else if ladder
		// as status is payment successful and no need to further check before returning order
	} else {
		http.Error(w, "cannot return an order which has not been paid for.", http.StatusBadRequest)
		return
	}

	var arg db.AddSavingsToWalletByUserIDParams
	arg.Savings = order.NetAmount
	arg.UserID = user.ID
	_, err = u.DB.AddSavingsToWalletByUserID(context.TODO(), arg)
	if err != nil {
		log.Warn("error adding return refund back to user wallet:", err.Error())
	} else {
		// edit payment status to be refunded
		var editPayArg db.EditPaymentStatusByOrderIDParams
		editPayArg.OrderID = order.ID
		editPayArg.Status = utils.StatusPaymentReturned
		_, err = u.DB.EditPaymentStatusByOrderID(context.TODO(), editPayArg)
		if err != nil {
			log.Warn("error changing payment status to returned in return order after successful adding money back to users's wallet")
		}

		// change order_item status to be returned
		orderItems, err := u.DB.GetOrderItemsByOrderID(context.TODO(), orderID)
		if err != nil {
			log.Warn("error fetching orderItems by orderId in ReturnOrderHandler")
		} else {
			// if there are no errors fetching order_items
			// then change the status of each orderItem
			// and increment the product back to the stock
			for _, v := range orderItems {
				var editOIArg db.EditOrderItemStatusByIDParams
				editOIArg.ID = v.ID
				editOIArg.Status = utils.StatusOrderReturned
				newOI, dbErr := u.DB.EditOrderItemStatusByID(context.TODO(), editOIArg)
				if dbErr != nil {
					log.Warn("error editing orderItem status in returnOrderHandler after returning payment:", dbErr.Error())
				} else {
					msg := fmt.Sprintf("changed oi status from %s to %s ", v.Status, newOI.Status)
					log.Info(msg)
				}

				// increment the products back to stock
				var incArg db.IncProductStockByIDParams
				incArg.IncQuantity = v.Quantity
				incArg.ProductID = v.ProductID
				product, dbErr := u.DB.IncProductStockByID(context.TODO(), incArg)
				if dbErr != nil {
					log.Warn("error incrementing stock after returning order in returnOrderHandler:", dbErr.Error())
				} else {
					msg := fmt.Sprintf("incremented product:%s of restock quantity: %d(added: %d)", product.Name, product.Stock, incArg.IncQuantity)
					log.Info(msg)
				}

			}
		}
	}

	msg := "successfully returned order"
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte(msg))
}

func (u *User) GetWishListHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	wishListItems, err := u.DB.GetAllWishListItemsWithProductNameByUserID(context.TODO(), user.ID)
	if err != nil {
		log.Error("error fetching wishlist items in GetWishListHandler:", err.Error())
		http.Error(w, "internal error fetching wishList items", http.StatusInternalServerError)
		return
	}

	type respWItem struct {
		ID          uuid.UUID `json:"id"`
		ProductID   uuid.UUID `json:"product_id"`
		ProductName string    `json:"product_name"`
	}
	var respWItems []respWItem
	for _, w := range wishListItems {
		var temp respWItem
		temp.ID = w.ID
		temp.ProductID = w.ProductID
		temp.ProductName = w.ProductName
		respWItems = append(respWItems, temp)
	}

	var resp struct {
		Data    []respWItem `json:"data"`
		Message string      `json:"message"`
	}
	resp.Data = respWItems
	resp.Message = "successfully fetched wishlist items"
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) AddProductToWishListHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	productIDStr := r.URL.Query().Get("product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		http.Error(w, "product_id not valid", http.StatusBadRequest)
		return
	}
	product, err := u.DB.GetProductByID(context.TODO(), productID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid product_id", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Error("error fetching product in ADdProdutToWishListHandler:", err.Error())
		http.Error(w, "internal error adding product to wishlist", http.StatusInternalServerError)
		return
	}

	// check if the product is already in the wishlist
	wishlistItem, err := u.DB.GetWishListItemByUserAndProductID(context.TODO(), db.GetWishListItemByUserAndProductIDParams{
		UserID:    user.ID,
		ProductID: product.ID,
	})
	if err == sql.ErrNoRows {
		// to make sure there are no duplicate items in wishlist
	} else if err != nil {
		log.Error("error fetching wishListItem in AddProductToWishListHandler:", err.Error())
		http.Error(w, "internal server error adding wishlistItem", http.StatusInternalServerError)
		return
	} else {
		msg := fmt.Sprintf("product: %s already exists in wishlist", product.Name)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	wishlistItem, err = u.DB.AddWishListItem(context.TODO(), db.AddWishListItemParams{
		UserID:    user.ID,
		ProductID: productID,
	})
	if err != nil {
		log.Error("error adding wishlistItem in ADdProdutToWishlistHandler:", err.Error())
		http.Error(w, "internal error adding wishlist item", http.StatusInternalServerError)
		return
	}

	var resp struct {
		ID          uuid.UUID `json:"id"`
		ProductID   uuid.UUID `josn:"product_id"`
		ProductName string    `json:"product_name"`
		Message     string    `json:"message"`
	}
	resp.ID = wishlistItem.ID
	resp.ProductID = wishlistItem.ProductID
	resp.ProductName = product.Name
	resp.Message = "successfully added wishlistItem"
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) RemoveWishListItemHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	productIDStr := r.URL.Query().Get("product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		http.Error(w, "product_id not valid", http.StatusBadRequest)
		return
	}
	product, err := u.DB.GetProductByID(context.TODO(), productID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid product_id", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Error("error fetching product in ADdProdutToWishListHandler:", err.Error())
		http.Error(w, "internal error adding product to wishlist", http.StatusInternalServerError)
		return
	}
	_, err = u.DB.GetWishListItemByUserAndProductID(context.TODO(), db.GetWishListItemByUserAndProductIDParams{
		UserID:    user.ID,
		ProductID: product.ID,
	})
	if err == sql.ErrNoRows {
		http.Error(w, "the product is not added in wishlist to delete the item", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Error("error fetching wishlistItem in RemoveWishListItemHandler:", err.Error())
		http.Error(w, "Internal error removing wishlist item", http.StatusInternalServerError)
		return
	}
	k, err := u.DB.DeleteWishListItemByUserAndProductID(context.TODO(), db.DeleteWishListItemByUserAndProductIDParams{
		UserID:    user.ID,
		ProductID: product.ID,
	})
	if err != nil {
		log.Error("error deleting wishlistItem in RemoveWishListItemHandler:", err.Error())
		http.Error(w, "internal errror removing wishlistItem", http.StatusInternalServerError)
		return
	} else if k == 0 {
		log.Error("error no rows affected after query execution in RemoveWishListItemHandler")
		http.Error(w, "internal errror removing wishlistItem", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	msg := "successfully deleted wishlistItem"
	w.Write([]byte(msg))
}

func (u *User) RemoveAllWishListHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	err := u.DB.DeleteAllWishListItemsByUserID(context.TODO(), user.ID)
	if err != nil {
		log.Error("error deleting all wishlist items for user in RemoveAllWishListHandler:", err.Error())
		http.Error(w, "internal error clearing all wishlist items", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	msg := "successfully deleted all wishlist items for the user"
	w.Write([]byte(msg))
}

func (u *User) AddWishListToCartHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	cartItems, err := u.DB.AddAllWishListItemsToCarts(context.TODO(), user.ID)
	if err != nil {
		log.Error("error adding wishListItems to cart in AddWishListToCartHandler:", err.Error())
		http.Error(w, "internal server error adding wishlist items to cart", http.StatusInternalServerError)
		return
	} else if len(cartItems) == 0 {
		http.Error(w, "no items in wishlist to add to cart", http.StatusBadRequest)
		return
	} else {
		err = u.DB.DeleteAllWishListItemsByUserID(context.TODO(), user.ID)
		if err != nil {
			log.Error("error removing wishlistItems after adding it to cart in AddWishListToCartHandler:", err.Error())
		}
	}

	type respCart struct {
		CartItemID uuid.UUID `json:"cart_item_id"`
		ProductID  uuid.UUID `json:"product_id"`
		Quantity   int       `json:"quantity"`
	}

	var respItems []respCart
	for _, ci := range cartItems {
		var temp respCart
		temp.CartItemID = ci.ID
		temp.ProductID = ci.ProductID
		temp.Quantity = int(ci.Quantity)
		respItems = append(respItems, temp)
	}

	var resp struct {
		Data    []respCart `json:"data"`
		Message string     `json:"message"`
	}

	resp.Data = respItems
	resp.Message = "successfully added wishlistItems to cart"

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
