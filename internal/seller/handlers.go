package seller

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/validators"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Seller struct {
	DB *db.Queries
}

func (s *Seller) OwnProductsHandler(w http.ResponseWriter, r *http.Request) {
	var user, ok = r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("user not found in request context")
		http.Error(w, "user not found in reqeust context", http.StatusInternalServerError)
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
	var user, ok = r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("user not found in request context")
		http.Error(w, "user not found in request context", http.StatusInternalServerError)
		return
	}
	var arg struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Price       float64  `json:"price"`
		Stock       int      `json:"stock"`
		Categories  []string `json:"categories"`
	}
	err := json.NewDecoder(r.Body).Decode(&arg)
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
	var user, ok = r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("user not found in request context")
		http.Error(w, "user not found in request context", http.StatusInternalServerError)
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
	var user, ok = r.Context().Value(utils.UserKey).(db.GetUserBySessionIDRow)
	if !ok {
		log.Warn("user not found in request context")
		http.Error(w, "user not found in request context", http.StatusInternalServerError)
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
