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
	}

	product, err := u.DB.GetProductByID(context.TODO(), req.ProductID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid productID", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "internal server error fetching product", http.StatusInternalServerError)
		return
	} else if req.Quantity <= 0 {
		http.Error(w, "invalid quantity", http.StatusBadRequest)
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
		resp.Message = "successfully added cart item"

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
	if cartItem.Quantity < product.Stock {
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
	}
	resp.Data = respEditItem
	resp.Message = "successfully updated cart item on adding the cart item on already existing cart item"
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
		var arg db.AddCartItemParams
		arg.UserID = user.ID
		arg.ProductID = req.ProductID
		if req.Quantity > int(product.Stock) {
			Err = append(Err, "adding more quantity of product than there is stock. Reverting the quantity back to maximum allotable")
			req.Quantity = int(product.Stock)
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
	if req.Quantity > int(product.Stock) {
		Err = append(Err, "edit cartItem with more quantity than there is stock. Reallocation the cartItem to the maximum possible")
		req.Quantity = int(product.Stock)
	} else {
		editArg.Quantity = int32(req.Quantity)
	}
	// edit the cartItem
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
	var req struct {
		ShippingAddressID uuid.UUID `json:"shipping_address_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request data", http.StatusBadRequest)
		return
	}

	// to add and display insignificant errors
	var Err []string
	// get a valid address for the shipping address
	address, err := u.DB.GetAddressByID(context.TODO(), req.ShippingAddressID)
	if err == sql.ErrNoRows {
		http.Error(w, "not a valid addressID", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Warn("error fetching address by id for order:", err.Error())
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

	// sum total for payment amount
	var sumTotal float64
	// add cartItems to orderItems
	for _, v := range cartItems {
		var addArg db.AddOrderITemParams
		addArg.OrderID = order.ID
		addArg.ProductID = v.ProductID
		addArg.Quantity = v.Quantity
		addArg.TotalAmount = v.TotalAmount
		sumTotal += v.TotalAmount
		_, err = u.DB.AddOrderITem(context.TODO(), addArg)
		if err != nil {
			log.Warn("error adding cartItem to order_item:", err.Error())
			http.Error(w, "internal error adding cartItem to order_items", http.StatusInternalServerError)
			return
		}
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
	payArg.OrderID = order.ID
	payArg.Method = "cod"
	payArg.Status = "processing"
	payArg.TotalAmount = sumTotal
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

	var resp struct {
		Message         string                         `json:"message"`
		Phone           int                            `json:"phone"`
		Payment         db.Payment                     `json:"payment"`
		OrderItems      []db.GetOrderItemsByOrderIDRow `json:"order_items"`
		ShippingAddress db.AddShippingAddressRow       `json:"shipping_address"`
		Err             []string                       `json:"error"`
	}
	resp.Phone = int(user.Phone.Int64)
	resp.Message = "successfully added the cart items to orders"
	resp.Payment = payment
	resp.OrderItems = orderItems
	resp.ShippingAddress = shipAddr
	resp.Err = Err
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *User) CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

}

func (u *User) ReturnOrderHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

}
