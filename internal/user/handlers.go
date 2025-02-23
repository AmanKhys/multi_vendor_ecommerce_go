package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
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
	} else if req.Quantity <= 0 {
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
	if req.Quantity > int(product.Stock) {
		Err = append(Err, "edit cartItem with more quantity than there is stock. Reallocation the cartItem to the maximum possible")
		editArg.Quantity = product.Stock
	} else {
		editArg.Quantity = int32(req.Quantity)
	}
	// edit the cartItem
	editedItem, err := u.DB.EditCartItemByID(context.TODO(), editArg)
	if err != nil {

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
		ProductID   uuid.UUID `json:"product_id"`
		ProductName string    `json:"product_name"`
		Price       float64   `json:"price"`
		Quantity    int       `json:"quantity"`
		TotalAmount float64   `json:"total_amount"`
	}
	// respOrder struct
	type respOrder struct {
		OrderID    uuid.UUID       `json:"order_id"`
		OrderItems []respOrderItem `json:"order_items"`
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
			for _, oi := range orderItems {
				var orderItem respOrderItem
				orderItem.OrderItemID = oi.ID
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
	// get a valid address for the shipping address
	address, err := u.DB.GetAddressByID(context.TODO(), ShippingAddressID)
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
		addArg.Price = v.Price
		addArg.Quantity = v.Quantity
		addArg.TotalAmount = v.TotalAmount
		sumTotal += v.TotalAmount
		_, err = u.DB.AddOrderITem(context.TODO(), addArg)
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

		// make Err slice to give for response errors
		var Err []string
		// decrement payment on cancelling order
		payment, err := u.DB.DecPaymentAmountByOrderItemID(context.TODO(), orderItem.ID)
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
			Message     string      `json:"message"`
			Err         []string    `json:"errors"`
			NewPayment  RespPayment `json:"new_payment"`
		}
		resp.OrderItemID = orderItem.ID
		resp.Message = "order_item has been cancelled."
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

	// cancel orderITems for the orderID
	err = u.DB.CancelOrderByID(context.TODO(), orderID)
	if err != nil {
		log.Warn("error cancelling orderItems by orderID in CancelOrderHandler:", err.Error())
		http.Error(w, "internal error cancelling orderItems by orderID", http.StatusInternalServerError)
		return
	} else {
		err = u.DB.CancelPaymentByOrderID(context.TODO(), orderID)
		if err != nil {
			log.Warn("error returning payment by orderID in CancelOrderHandler:", err.Error())
			http.Error(w, "intenral error cancelling payment by orderID after successfully cancelling orderItems", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	message := "successfully cancelled order and the payment"
	w.Write([]byte(message))
}

func (u *User) ReturnOrderHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

}
