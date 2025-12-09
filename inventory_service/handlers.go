package inventoryservice

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	db "inventory_service/db/sqlc"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/helpers"
	middleware "github.com/amankhys/multi_vendor_ecommerce_go/pkg/middlewares"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/validators"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository"
	"github.com/google/uuid"
)

var dbConn = repository.NewDBConfig("user")
var DB = db.New(dbConn)
var u = User{DB: DB}
var helper = helpers.Helper{
	DB: DB,
}

func RegisterRoutes(mux *http.ServeMux) {
	// user side
	mux.HandleFunc("GET /user/products", u.ProductsHandler)
	mux.HandleFunc("GET /user/product", u.ProductHandler)
	mux.HandleFunc("POST /user/product/review", middleware.AuthenticateUserMiddleware(u.AddProductReviewHandler, utils.UserRole))
	mux.HandleFunc("GET /user/category", u.CategoryHandler)
	mux.HandleFunc("GET /user/wishlist", middleware.AuthenticateUserMiddleware(u.GetWishListHandler, utils.UserRole))
	mux.HandleFunc("POST /user/wishlist/add", middleware.AuthenticateUserMiddleware(u.AddProductToWishListHandler, utils.UserRole))
	mux.HandleFunc("DELETE /user/wishlist/item/delete", middleware.AuthenticateUserMiddleware(u.RemoveWishListItemHandler, utils.UserRole))
	mux.HandleFunc("DELETE /user/wishlist/delete", middleware.AuthenticateUserMiddleware(u.RemoveAllWishListHandler, utils.UserRole))
	mux.HandleFunc("POST /user/wishlist/add_to_cart", middleware.AuthenticateUserMiddleware(u.AddWishListToCartHandler, utils.UserRole))

	// seller side
	s := &Seller{DB: DB}
	mux.HandleFunc("GET /seller/products", middleware.AuthenticateUserMiddleware(s.OwnProductsHandler, utils.SellerRole))
	mux.HandleFunc("GET /seller/product", middleware.AuthenticateUserMiddleware(s.ProductDetailsHandler, utils.SellerRole))
	mux.HandleFunc("POST /seller/product/add", middleware.AuthenticateUserMiddleware(s.AddProductHandler, utils.SellerRole))
	mux.HandleFunc("PUT /seller/product/edit", middleware.AuthenticateUserMiddleware(s.EditProductHandler, utils.SellerRole))
	mux.HandleFunc("DELETE /seller/product/delete", middleware.AuthenticateUserMiddleware(s.DeleteProductHandler, utils.SellerRole))

	mux.HandleFunc("GET /seller/categories", middleware.AuthenticateUserMiddleware(s.GetAllCategoriesHandler, utils.SellerRole))
	mux.HandleFunc("POST /seller/category/add", middleware.AuthenticateUserMiddleware(s.AddProductToCategoryHandler, utils.SellerRole))

	// admin side
	a := &Admin{DB: DB}
	mux.HandleFunc("GET /admin/products", middleware.AuthenticateUserMiddleware(a.AdminProductsHandler, utils.AdminRole))
	mux.HandleFunc("DELETE /admin/product/delete", middleware.AuthenticateUserMiddleware(a.DeleteProductHandler, utils.AdminRole))

	mux.HandleFunc("GET /admin/categories", middleware.AuthenticateUserMiddleware(a.AdminCategoriesHandler, utils.AdminRole))
	mux.HandleFunc("POST /admin/category/add", middleware.AuthenticateUserMiddleware(a.AddCategoryHandler, utils.AdminRole))
	mux.HandleFunc("PUT /admin/category/edit", middleware.AuthenticateUserMiddleware(a.EditCategoryHandler, utils.AdminRole))
	mux.HandleFunc("DELETE /admin/category/delete", middleware.AuthenticateUserMiddleware(a.DeleteCategoryHandler, utils.AdminRole))

}

type User struct{ DB *db.Queries }

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
	type respProduct struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Price       float64   `json:"price"`
		Stock       int32     `json:"stock"`
		SellerID    uuid.UUID `json:"seller_id"`
	}

	var respProductsData []respProduct

	for _, p := range products {
		var temp respProduct
		temp.ID = p.ID
		temp.Name = p.Name
		temp.Description = p.Description
		temp.Price = p.Price
		temp.Stock = p.Stock
		temp.SellerID = p.SellerID
		respProductsData = append(respProductsData, temp)
	}

	var resp struct {
		Data    []respProduct `json:"data"`
		Message string        `json:"message"`
	}
	resp.Data = respProductsData
	resp.Message = "successfully fetched products from category:" + categoryName
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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

type Seller struct{ DB *db.Queries }

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
	type respProduct struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Price       float64   `json:"price"`
		Stock       int32     `json:"stock"`
		SellerID    uuid.UUID `json:"seller_id"`
	}

	var respProductsData []respProduct

	for _, p := range products {
		var temp respProduct
		temp.ID = p.ID
		temp.Name = p.Name
		temp.Description = p.Description
		temp.Price = p.Price
		temp.Stock = p.Stock
		temp.SellerID = p.SellerID

		respProductsData = append(respProductsData, temp)
	}

	w.Header().Set("Content-Type", "application/json")
	var resp struct {
		Data []respProduct `json:"data"`
		Err  []string      `json:"errors"`
	}
	resp.Data = respProductsData
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
	type respProduct struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Price       float64   `json:"price"`
		Stock       int32     `json:"stock"`
		SellerID    uuid.UUID `json:"seller_id"`
	}

	var respProductData = respProduct{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		SellerID:    product.SellerID,
	}
	var resp struct {
		Data       respProduct `json:"data"`
		Message    string      `json:"message"`
		Categories []string    `json:"categories"`
		Err        []string    `json:"errors"`
	}
	resp.Data = respProductData
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

	type respProduct struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Price       float64   `json:"price"`
		Stock       int32     `json:"stock"`
		SellerID    uuid.UUID `json:"seller_id"`
	}

	var respProductData = respProduct{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		SellerID:    product.SellerID,
	}
	var resp struct {
		Data            respProduct `json:"data"`
		Message         string      `json:"message"`
		CategoriesAdded []string    `json:"categories_added"`
		Err             []string    `json:"error"`
	}
	resp.Data = respProductData
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

	type respProduct struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Price       float64   `json:"price"`
		Stock       int32     `json:"stock"`
		SellerID    uuid.UUID `json:"seller_id"`
	}

	var respProductData = respProduct{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		SellerID:    product.SellerID,
	}
	var resp struct {
		Data            respProduct `json:"data"`
		Message         string      `json:"message"`
		Err             []string    `json:"error"`
		CategoriesAdded []string    `json:"categories_added"`
	}
	resp.Err = Err
	resp.CategoriesAdded = CategoriesAdded
	resp.Data = respProductData
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
	type respProduct struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Price       float64   `json:"price"`
		Stock       int32     `json:"stock"`
		SellerID    uuid.UUID `json:"seller_id"`
	}

	var respProductData = respProduct{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		SellerID:    product.SellerID,
	}
	var resp struct {
		Product respProduct `json:"product"`
		Message string      `json:"message"`
	}
	resp.Product = respProductData
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
	type respCategory struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	var respCategories []respCategory

	for _, c := range categories {
		var temp respCategory
		temp.ID = c.ID
		temp.Name = c.Name
		respCategories = append(respCategories, temp)
	}
	var resp struct {
		Data    []respCategory `json:"data"`
		Message string         `json:"message"`
	}
	resp.Data = respCategories
	resp.Message = "successfully fetched all categories"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Seller) AddProductToCategoryHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
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
	type respProduct struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Price       float64   `json:"price"`
		Stock       int32     `json:"stock"`
		SellerID    uuid.UUID `json:"seller_id"`
	}

	var respProductData = respProduct{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		SellerID:    product.SellerID,
	}
	var resp struct {
		Product      respProduct `json:"product"`
		CategoryName string      `json:"category_name"`
		Message      string      `json:"message"`
	}
	resp.Product = respProductData
	resp.CategoryName = category.Name
	resp.Message = "successfully added product to category items"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type Admin struct{ DB *db.Queries }

func (a *Admin) AdminProductsHandler(w http.ResponseWriter, r *http.Request) {
	type respProduct struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Price       float64   `json:"price"`
		Stock       int32     `json:"stock"`
		SellerID    uuid.UUID `json:"seller_id"`
		IsDeleted   bool      `json:"is_deleted"`
	}

	var respProducts []respProduct
	var resp struct {
		Data    []respProduct `json:"data"`
		Message string        `json:"message"`
	}
	products, err := a.DB.GetAllProductsForAdmin(context.TODO())
	if err != nil {
		log.Warn(err)
		http.Error(w, fmt.Errorf("failed : %w", err).Error(), http.StatusBadRequest)
		return
	}

	for _, p := range products {
		var temp respProduct
		temp.ID = p.ID
		temp.Name = p.Name
		temp.Description = p.Description
		temp.Price = p.Price
		temp.Stock = p.Stock
		temp.SellerID = p.SellerID
		temp.IsDeleted = p.IsDeleted

		respProducts = append(respProducts, temp)
	}

	resp.Data = respProducts
	resp.Message = "successfully fetched all products"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}

func (a *Admin) AdminCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	type respCategory struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	var respCateogies []respCategory
	var resp struct {
		Data    []respCategory `json:"data"`
		Message string         `json:"message"`
	}
	categories, err := a.DB.GetAllCategoriesForAdmin(context.TODO())
	if err != nil {
		log.Warn(err)
		http.Error(w, fmt.Errorf("failed : %w", err).Error(), http.StatusBadRequest)
		return
	}

	for _, c := range categories {
		var temp respCategory
		temp.ID = c.ID
		temp.Name = c.Name
		respCateogies = append(respCateogies, temp)
	}

	resp.Data = respCateogies
	resp.Message = "successfully fetched all categories"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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
