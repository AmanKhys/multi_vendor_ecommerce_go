package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type User struct {
	DB *db.Queries
}

// //////////////////////////////////
// product handlers

func (u *User) ProductsHandler(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Data []db.Product `json:"data"`
	}
	products, err := u.DB.GetAllProducts(context.TODO())
	if err != nil {
		log.Warn("couldn't fetch from products table: ", err)
		http.Error(w, "internal server error: couldn't fetch products", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	resp := Response{Data: products}
	json.NewEncoder(w).Encode(resp)
}

func (u *User) ProductHandler(w http.ResponseWriter, r *http.Request) {
	ProductID := r.URL.Query().Get("id")
	id, err := uuid.Parse(ProductID)
	if err != nil {
		http.Error(w, "wrong productID format", http.StatusBadRequest)
		return
	}
	product, err := u.DB.GetProductByID(context.TODO(), id)
	if err != nil {
		http.Error(w, "no such product exists", http.StatusNotFound)
		return
	}
	type Response struct {
		Data db.Product `json:"data"`
	}
	w.Header().Set("Content-Type", "application/json")
	resp := Response{Data: product}
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
	}

	product, err := u.DB.GetProductByID(context.TODO(), req.ProductID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid productID", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "internal server erro fetching product", http.StatusInternalServerError)
		return
	} else if req.Quantity <= 0 {
		http.Error(w, "invalid quantity", http.StatusBadRequest)
		return
	}

	var getArg db.GetCartItemByUserIDAndProductIDParams
	getArg.ProductID = product.ID
	getArg.UserID = user.ID
	cartItem, err := u.DB.GetCartItemByUserIDAndProductID(context.TODO(), getArg)
	// add cartItem if carts doesn't already have the particular combination of cartItem
	if err == sql.ErrNoRows {
		var arg db.AddCartItemParams
		arg.UserID = user.ID
		arg.ProductID = req.ProductID
		arg.Quantity = int32(req.Quantity)
		arg.ProductID = product.ID
		arg.UserID = user.ID
		item, err := u.DB.AddCartItem(context.TODO(), arg)
		if err != nil {
			log.Warn("internal error adding cartItem:", err.Error())
			http.Error(w, "internal error adding cartItem", http.StatusInternalServerError)
			return
		}
		var respItem = respCartItem{
			CartID:      cartItem.ID,
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    int(item.Quantity),
		}
		var resp struct {
			Data    respCartItem `json:"data"`
			Message string       `json:"message"`
		}
		resp.Data = respItem
		resp.Message = "successfully added cart item"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return // make sure to return after sending the response
	} else if err != nil {
		log.Error(w, "internal error fetching cartItem to check before AddCart:", err.Error())
		http.Error(w, "internal error fetching cartItem to check before AddCart", http.StatusInternalServerError)
		return
	}

	// update existing cartItem if there is a valid cartItem that is fetched from the database
	var editArg db.EditCartItemByIDParams
	editArg.ID = cartItem.ID
	editArg.Quantity = cartItem.Quantity + 1
	editItem, err := u.DB.EditCartItemByID(context.TODO(), editArg)
	if err != nil {
		log.Warn("internal error editing cartItem on adding cartItem on already existing item", err.Error())
		http.Error(w, "internal error updating cartItem on adding the cartItem again.", http.StatusInternalServerError)
		return
	}

	var respEditITem respCartItem
	respEditITem.ProductName = product.Name
	respEditITem.Quantity = int(editItem.Quantity)
	var resp struct {
		Data    respCartItem `json:"data"`
		Message string       `json:"message"`
	}
	resp.Data = respEditITem
	resp.Message = "successfully updated cart item on adding the cart item on already existing cart item"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) EditCartHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
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
	}

	product, err := u.DB.GetProductByID(context.TODO(), req.ProductID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid productID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("internal error fetching product from cartID:", err.Error())
		http.Error(w, "internal server error fetching product from cartID", http.StatusInternalServerError)
		return
	}

	var getArg db.GetCartItemByUserIDAndProductIDParams
	getArg.ProductID = product.ID
	getArg.UserID = user.ID
	cartItem, err := u.DB.GetCartItemByUserIDAndProductID(context.TODO(), getArg)
	// add product if the product is not in carts;
	if err == sql.ErrNoRows {
		var arg db.AddCartItemParams
		arg.UserID = user.ID
		arg.ProductID = req.ProductID
		arg.Quantity = int32(req.Quantity)
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
		}
		resp.Data = respItem
		resp.Message = "successfully added cart item since already didn't exist."

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return

	} else if err != nil {
		log.Warn("error fetching cartItem before editing cartItem using GetCartItemByID:", err.Error())
		http.Error(w, "internal error fetching cart item before editing", http.StatusInternalServerError)
		return
	}

	// edit cart item when there is an existing cart item for the product
	var editArg db.EditCartItemByIDParams
	editArg.ID = cartItem.ID
	editArg.Quantity = int32(req.Quantity)
	editedItem, err := u.DB.EditCartItemByID(context.TODO(), editArg)
	if err != nil {
		log.Warn("internal error editing cartItem:", err.Error())
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
	}
	resp.Data = respItem
	resp.Message = "successfully edited cartItem"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) DeleteCartHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}
	var req struct {
		ProductID uuid.UUID `json:"product_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request data", http.StatusBadRequest)
		return
	}
	product, err := u.DB.GetProductByID(context.TODO(), req.ProductID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid productID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("internal erro fetching product to delete it from cart:", err.Error())
		http.Error(w, "internal error fetching product to delete it from cart", http.StatusInternalServerError)
		return
	}
	var deleteArg db.DeleteCartItemByUserIDAndProductIDParams
	deleteArg.ProductID = product.ID
	deleteArg.UserID = user.ID
	k, err := u.DB.DeleteCartItemByUserIDAndProductID(context.TODO(), deleteArg)
	if err != nil {
		log.Warn("internal error deleting cartItem with valid productID:", err.Error())
		http.Error(w, "internal error deleting cartItem", http.StatusInternalServerError)
		return
	} else if k == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "text/plain")
		message := "there was no cartItem with the said productID. No cartItem deleted"
		w.Write([]byte(message))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	message := fmt.Sprintf("product: %s deleted successfully from cartItems", product.Name)
	w.Write([]byte(message))
}
