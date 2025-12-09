package payment_service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	db "payment_service/db/sqlc"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/chartGen"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/envname"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/helpers"
	middleware "github.com/amankhys/multi_vendor_ecommerce_go/pkg/middlewares"
	paymenthelper "github.com/amankhys/multi_vendor_ecommerce_go/pkg/payment"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/validators"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
)

var dbConn = db.NewDBConfig("user")
var DB = db.New(dbConn)
var u = User{DB: DB}
var helper = helpers.Helper{
	DB: DB,
}

func RegisterRoutes(mux *http.ServeMux) {

	mux.HandleFunc("GET /user/cart", middleware.AuthenticateUserMiddleware(u.GetCartHandler, utils.UserRole))
	mux.HandleFunc("POST /user/cart/add", middleware.AuthenticateUserMiddleware(u.AddCartHandler, utils.UserRole))
	mux.HandleFunc("PUT /user/cart/edit", middleware.AuthenticateUserMiddleware(u.EditCartHandler, utils.UserRole))
	mux.HandleFunc("DELETE /user/cart/delete", middleware.AuthenticateUserMiddleware(u.DeleteCartHandler, utils.UserRole))

	mux.HandleFunc("GET /user/orders", middleware.AuthenticateUserMiddleware(u.GetOrdersHandler, utils.UserRole))
	mux.HandleFunc("GET /user/orders/items", middleware.AuthenticateUserMiddleware(u.GetOrderItemsHandler, utils.UserRole))
	mux.HandleFunc("POST /user/orders/create", middleware.AuthenticateUserMiddleware(u.AddCartToOrderHandler, utils.UserRole))
	mux.HandleFunc("PUT /user/orders/cancel", middleware.AuthenticateUserMiddleware(u.CancelOrderHandler, utils.UserRole))
	mux.HandleFunc("PUT /user/orders/item/cancel", middleware.AuthenticateUserMiddleware(u.CancelOrderItemHandler, utils.UserRole))
	mux.HandleFunc("PUT /user/orders/return", middleware.AuthenticateUserMiddleware(u.ReturnOrderHandler, utils.UserRole))

	mux.HandleFunc("GET /user/orders/makepayment", middleware.AuthenticateUserMiddleware(u.MakeOnlinePaymentHandler, utils.UserRole))
	mux.HandleFunc("POST /user/orders/makepayment/success", middleware.AuthenticateUserMiddleware(u.PaymentSuccessHandler, utils.UserRole))
	mux.HandleFunc("GET /user/orders/invoice", middleware.AuthenticateUserMiddleware(u.InvoiceHandler, utils.UserRole))

	s := &Seller{DB: DB}
	mux.HandleFunc("GET /seller/orders", middleware.AuthenticateUserMiddleware(s.GetOrdersHandler, utils.SellerRole))
	mux.HandleFunc("PUT /seller/orders/status", middleware.AuthenticateUserMiddleware(s.ChangeOrderStatusHandler, utils.SellerRole))
	mux.HandleFunc("GET /seller/sales_report", middleware.AuthenticateUserMiddleware(s.SalesReportHandler, utils.SellerRole))

	a := &Admin{DB: DB}
	mux.HandleFunc("GET /admin/orders", middleware.AuthenticateUserMiddleware(a.GetOrderItemsHandler, utils.AdminRole))
	mux.HandleFunc("PUT /admin/orders/deliver", middleware.AuthenticateUserMiddleware(a.DeliverOrderItemHandler, utils.AdminRole))

	mux.HandleFunc("GET /admin/coupons", middleware.AuthenticateUserMiddleware(a.AdminCouponsHandler, utils.AdminRole))
	mux.HandleFunc("POST /admin/coupons/add", middleware.AuthenticateUserMiddleware(a.AddCouponHandler, utils.AdminRole))
	mux.HandleFunc("PUT /admin/coupons/edit", middleware.AuthenticateUserMiddleware(a.EditCouponHandler, utils.AdminRole))
	mux.HandleFunc("DELETE /admin/coupons/delete", middleware.AuthenticateUserMiddleware(a.DeleteCouponHandler, utils.AdminRole))

	mux.HandleFunc("GET /admin/sales_report", middleware.AuthenticateUserMiddleware(a.SalesReportHandler, utils.AdminRole))
}

type User struct{ DB *db.Queries }

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

	type respCartItems struct {
		CartID      uuid.UUID `json:"cart_id"`
		ProductID   uuid.UUID `json:"product_id"`
		ProductName string    `json:"product_name"`
		Quantity    int32     `json:"quantity"`
		Price       float64   `json:"price"`
		TotalAmount float64   `json:"total_amount"`
	}

	var respCartItemsData []respCartItems
	var cartTotal float64
	for _, ci := range cartItems {
		var temp respCartItems
		temp.CartID = ci.CartID
		temp.ProductID = ci.ProductID
		temp.ProductName = ci.ProductName
		temp.Quantity = ci.Quantity
		temp.Price = ci.Price
		temp.TotalAmount = ci.TotalAmount

		cartTotal += ci.TotalAmount
		respCartItemsData = append(respCartItemsData, temp)
	}

	var resp struct {
		Data      []respCartItems `json:"data"`
		CartTotal float64         `json:"cart_total"`
		Message   string          `json:"message"`
	}
	resp.Data = respCartItemsData
	resp.CartTotal = cartTotal
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
		OrderDate     time.Time       `json:"order_date"`
		PaymentMethod string          `json:"payment_method"`
		PaymentStatus string          `json:"payment_status"`
		OrderItems    []respOrderItem `json:"order_items"`
	}

	// var respOrders, errors
	var respOrders []respOrder
	var Err []string

	for _, o := range orders {
		var temp respOrder
		// add order_date
		temp.OrderDate = o.CreatedAt
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
		OrderID     uuid.UUID `json:"order_id"`
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
		temp.OrderID = v.OrderID
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

	// for future calculations
	var discountAmount float64
	var ifCouponValid bool

	totalAmount, err := u.DB.GetSumOfCartItemsByUserID(context.TODO(), user.ID)
	if err != nil {
		log.Error("error fetching totalAmount for user in AddCartToOrderHandler:", err.Error())
		http.Error(w, "internal error fetching necessary data", http.StatusInternalServerError)
		return
	}

	// check whether payment is cod and total_amount > 1000
	// take the payment method from url query
	paymentMethod := r.URL.Query().Get("payment_method")
	if !(paymentMethod == utils.StatusPaymentMethodCod || paymentMethod == utils.StatusPaymentMethodRpay ||
		paymentMethod == utils.StatusPaymentMethodWallet) {
		http.Error(w, "invalid payment method", http.StatusBadRequest)
		return
	}
	if paymentMethod == utils.StatusPaymentMethodCod && totalAmount >= 1000 {
		http.Error(w, "cannot create order costing more than 1000rs on Cash On Delivery", http.StatusBadRequest)
		return
	} else if paymentMethod == utils.StatusPaymentMethodWallet {
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

	// set value for discountAmount
	// also set ifCouponValid true if coupon exists
	if ifCouponExists && coupon.TriggerPrice <= totalAmount &&
		time.Now().After(coupon.StartDate) && time.Now().Before(coupon.EndDate) && !coupon.IsDeleted {
		ifCouponValid = true
		if coupon.DiscountType == utils.CouponDiscountTypeFlat {
			discountAmount = coupon.DiscountAmount
		} else if coupon.DiscountType == utils.CouponDiscountTypePercentage {
			discountAmount = coupon.DiscountAmount * totalAmount / 100
		}
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

	type respOrder struct {
		ID             uuid.UUID     `json:"id"`
		UserID         uuid.UUID     `json:"user_id"`
		TotalAmount    float64       `json:"total_amount"`
		CouponID       uuid.NullUUID `json:"coupon_id"`
		DiscountAmount float64       `json:"discount_amount"`
		NetAmount      float64       `json:"net_amount"`
		OrderDate      time.Time     `json:"created_at"`
	}
	var respOrderData = respOrder{
		ID:             order.ID,
		UserID:         order.UserID,
		TotalAmount:    order.TotalAmount,
		CouponID:       order.CouponID,
		DiscountAmount: order.DiscountAmount,
		NetAmount:      order.NetAmount,
		OrderDate:      order.CreatedAt,
	}

	type respOrderItem struct {
		ID          uuid.UUID `json:"id"`
		ProductID   uuid.UUID `json:"product_id"`
		Price       float64   `json:"price"`
		Quantity    int32     `json:"quantity"`
		TotalAmount float64   `json:"total_amount"`
		Status      string    `json:"status"`
		ProductName string    `json:"product_name"`
	}

	var respOrderItemsData []respOrderItem
	for _, oi := range orderItems {
		var temp respOrderItem
		temp.ID = oi.ID
		temp.ProductID = oi.ProductID
		temp.ProductName = oi.ProductName
		temp.Price = oi.Price
		temp.Quantity = oi.Quantity
		temp.TotalAmount = oi.TotalAmount
		temp.Status = oi.Status

		respOrderItemsData = append(respOrderItemsData, temp)
	}

	type respShippingAddress struct {
		ID         uuid.UUID `json:"id"`
		HouseName  string    `json:"house_name"`
		StreetName string    `json:"street_name"`
		Town       string    `json:"town"`
		District   string    `json:"district"`
		State      string    `json:"state"`
		Pincode    int32     `json:"pincode"`
	}

	var respShipAddrData = respShippingAddress{
		ID:         shipAddr.ID,
		HouseName:  shipAddr.HouseName,
		StreetName: shipAddr.StreetName,
		Town:       shipAddr.Town,
		District:   shipAddr.District,
		State:      shipAddr.State,
		Pincode:    shipAddr.Pincode,
	}
	type respPayment struct {
		ID            uuid.UUID      `json:"id"`
		Method        string         `json:"method"`
		Status        string         `json:"status"`
		TotalAmount   float64        `json:"total_amount"`
		TransactionID sql.NullString `json:"transaction_id"`
		CreatedAt     time.Time      `json:"created_at"`
	}

	var respPaymentData = respPayment{
		ID:            payment.ID,
		Method:        payment.Method,
		Status:        payment.Status,
		TotalAmount:   payment.TotalAmount,
		TransactionID: payment.TransactionID,
		CreatedAt:     payment.CreatedAt,
	}

	Messages = append(Messages, "successfully added the cart items to orders")
	var resp struct {
		Order           respOrder           `json:"order"`
		Phone           int                 `json:"phone"`
		Payment         respPayment         `json:"payment"`
		OrderItems      []respOrderItem     `json:"order_items"`
		ShippingAddress respShippingAddress `json:"shipping_address"`
		Err             []string            `json:"error"`
		Messages        []string            `json:"messages"`
	}

	resp.Phone = int(user.Phone.Int64)
	resp.Order = respOrderData
	resp.Payment = respPaymentData
	resp.OrderItems = respOrderItemsData
	resp.ShippingAddress = respShipAddrData
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
	rpOrderIDStr, err := paymenthelper.ExecuteRazorpay(payment.TotalAmount)
	if err != nil {
		log.Warn("error executing razorpay")
		http.Error(w, "internal error executing razorpay", http.StatusInternalServerError)
		return
	}

	RPayKey := os.Getenv(envname.RPID)
	// data for the razorpaytemplate
	data := RazorpayData{
		Key:           RPayKey,
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

	if paymenthelper.VerifyRazorpaySignature(resp.OrderID, resp.PaymentID, resp.Signature) {
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

func (u *User) InvoiceHandler(w http.ResponseWriter, r *http.Request) {
	user := helper.GetUserHelper(w, r)
	if user.ID == uuid.Nil {
		return
	}

	orderIDStr := r.URL.Query().Get("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		http.Error(w, "invalid order_id", http.StatusBadRequest)
		return
	}
	order, err := u.DB.GetOrderByID(context.TODO(), orderID)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid order_id", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Error("error fetching order in InvoiceHandler")
		http.Error(w, "internal error fetching order", http.StatusInternalServerError)
		return
	} else if order.UserID != user.ID {
		http.Error(w, "not user's order. Forbidden", http.StatusForbidden)
		return
	}

	payment, err := u.DB.GetPaymentByOrderID(context.TODO(), order.ID)
	if err != nil {
		log.Error("error fetching payment by Order in InvoiceHandler")
		http.Error(w, "internal error producing invoice", http.StatusInternalServerError)
		return
	} else if payment.Status == utils.StatusPaymentReturned {
		http.Error(w, "order returned", http.StatusBadRequest)
		return
	} else if payment.Status == utils.StatusPaymentCancelled {
		http.Error(w, "order cancelled", http.StatusBadRequest)
		return
	} else if payment.Status != utils.StatusPaymentSuccessful {
		http.Error(w, "order not paid", http.StatusBadRequest)
		return
	}

	orderItems, err := u.DB.GetOrderItemsByOrderID(context.TODO(), order.ID)
	if err != nil {
		log.Error("error fetching order items in InvoiceHandler")
		http.Error(w, "internal error producing invoice", http.StatusInternalServerError)
		return
	}

	shippingAddress, err := u.DB.GetShippingAddressByOrderID(context.TODO(), order.ID)
	if err != nil {
		log.Error("error fetching shipping address by orderID in InvoiceHandler")
		http.Error(w, "internal error fetching order invoice", http.StatusInternalServerError)
		return
	}

	type respData struct {
		OrderItemID uuid.UUID
		ProductName string
		ProductID   uuid.UUID
		SellerName  string
		Price       float64
		Quantity    int
		TaxAmount   float64
		NetAmount   float64
		TotalAmount float64
	}
	var resp []respData
	for _, oi := range orderItems {
		seller, err := u.DB.GetSellerByProductID(context.TODO(), oi.ProductID)
		if err != nil {
			log.Error("error fetching seller in InvoiceHandler")
			http.Error(w, "internal error producing invoice", http.StatusInternalServerError)
			return
		}
		tax := oi.TotalAmount * utils.OrderTaxPercentage
		resp = append(resp, respData{
			OrderItemID: oi.ID,
			ProductName: oi.ProductName,
			ProductID:   oi.ProductID,
			SellerName:  seller.Name,
			Price:       oi.Price,
			Quantity:    int(oi.Quantity),
			TaxAmount:   tax,
			NetAmount:   oi.TotalAmount - tax,
			TotalAmount: oi.TotalAmount,
		})
	}
	// check whether coupon is applied and add coupon and discount details to the pdf if there are any
	type Discount struct {
		CouponName    string  `json:"coupon_name"`
		DiscountType  string  `json:"discount_type"`
		TotalDiscount float64 `json:"discount"`
	}
	var orderDiscount Discount
	if order.CouponID.Valid {
		coupon, err := u.DB.GetCouponByID(context.TODO(), order.CouponID.UUID)
		if err == sql.ErrNoRows {
			log.Error("no coupon found for the order with couponID present in Invoice Handler")
		}
		orderDiscount = Discount{
			CouponName:    coupon.Name,
			DiscountType:  coupon.DiscountType,
			TotalDiscount: order.DiscountAmount,
		}
	}

	// Begin PDF generation
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Font setup
	pdf.SetFont("Arial", "", 12)
	pdf.SetAutoPageBreak(true, 10)

	// Header with branding
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(0, 10, "Toy Stores Ecom")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 8, "Email: toystores@gmail.com")
	pdf.Ln(4)
	pdf.Cell(0, 8, fmt.Sprintf("Invoice Date: %s", time.Now().Format("02 Jan 2006")))
	pdf.Ln(6)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(6)

	// Order Info
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, fmt.Sprintf("Invoice for Order ID: %s", order.ID))
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 8, fmt.Sprintf("Payment Status: %s", payment.Status))
	pdf.Ln(12)

	// Customer Details
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Customer Information:")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 6, fmt.Sprintf("Name: %s", user.Name))
	pdf.Ln(5)
	pdf.Cell(0, 6, fmt.Sprintf("Email: %s", user.Email))
	pdf.Ln(5)
	if user.Phone.Valid {
		pdf.Cell(0, 6, fmt.Sprintf("Phone: %d", user.Phone.Int64))
	} else {
		pdf.Cell(0, 6, "Phone: N/A")
	}
	pdf.Ln(10)

	// Shipping Address
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Shipping Address:")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 11)
	address := fmt.Sprintf("%s, %s, %s, %s, %s - %d",
		shippingAddress.HouseName, shippingAddress.StreetName, shippingAddress.Town,
		shippingAddress.District, shippingAddress.State, shippingAddress.Pincode)
	pdf.MultiCell(0, 6, address, "", "", false)
	pdf.Ln(6)

	// Payment Details
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Payment Details:")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 6, fmt.Sprintf("Method: %s", payment.Method))
	pdf.Ln(5)
	if payment.TransactionID.Valid {
		pdf.Cell(0, 6, fmt.Sprintf("Transaction ID: %s", payment.TransactionID.String))
	} else {
		pdf.Cell(0, 6, "Transaction ID: N/A")
	}
	pdf.Ln(12)

	// Discount / Coupon Info
	pdf.Ln(6)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Discount Details:")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 11)
	if order.CouponID.Valid && orderDiscount.CouponName != "" {
		pdf.Cell(0, 6, fmt.Sprintf("Coupon Code: %s", orderDiscount.CouponName))
		pdf.Ln(5)
		pdf.Cell(0, 6, fmt.Sprintf("Discount Type: %s", orderDiscount.DiscountType))
		pdf.Ln(5)
		pdf.Cell(0, 6, fmt.Sprintf("Total Discount: ₹%.2f", orderDiscount.TotalDiscount))
	} else {
		pdf.Cell(0, 6, "No coupon was applied for this order.")
	}
	pdf.Ln(12)

	// Table: Order Items
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Order Items")
	pdf.Ln(8)

	// Table Headers
	headers := []string{"Product", "Seller", "Price", "Qty", "Tax", "Net", "Total"}
	widths := []float64{40, 35, 20, 15, 20, 25, 25}

	pdf.SetFillColor(230, 230, 230)
	pdf.SetFont("Arial", "B", 11)
	for i, h := range headers {
		pdf.CellFormat(widths[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table Rows
	pdf.SetFont("Arial", "", 10)
	fill := false
	for _, item := range resp {
		pdf.SetFillColor(245, 245, 245)
		pdf.CellFormat(widths[0], 8, item.ProductName, "1", 0, "", fill, 0, "")
		pdf.CellFormat(widths[1], 8, item.SellerName, "1", 0, "", fill, 0, "")
		pdf.CellFormat(widths[2], 8, fmt.Sprintf("%.2f", item.Price), "1", 0, "R", fill, 0, "")
		pdf.CellFormat(widths[3], 8, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(widths[4], 8, fmt.Sprintf("%.2f", item.TaxAmount), "1", 0, "R", fill, 0, "")
		pdf.CellFormat(widths[5], 8, fmt.Sprintf("%.2f", item.NetAmount), "1", 0, "R", fill, 0, "")
		pdf.CellFormat(widths[6], 8, fmt.Sprintf("%.2f", item.TotalAmount), "1", 0, "R", fill, 0, "")
		pdf.Ln(-1)
		fill = !fill
	}

	// Summary
	pdf.Ln(8)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(130, 8, "Subtotal:")
	pdf.Cell(40, 8, fmt.Sprintf("%.2f", order.TotalAmount))
	pdf.Ln(6)

	if order.DiscountAmount > 0 {
		pdf.Cell(130, 8, "Discount:")
		pdf.Cell(40, 8, fmt.Sprintf("-%.2f", order.DiscountAmount))
		pdf.Ln(6)
	}

	pdf.Cell(130, 8, "Total Paid:")
	pdf.Cell(40, 8, fmt.Sprintf("%.2f", order.NetAmount))
	pdf.Ln(10)

	// Footer
	pdf.SetY(-20)
	pdf.SetFont("Arial", "I", 9)
	pdf.Cell(0, 10, "Thank you for shopping with Toy Stores Ecom!")

	// Output PDF
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		http.Error(w, "failed to generate invoice PDF", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=invoice-%s.pdf", order.ID.String()))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())

}

// seller side
type Seller struct{ DB *db.Queries }

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
		OrderDate   time.Time `json:"order_date"`
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
		temp.OrderDate = v.CreatedAt
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

// admin side
type Admin struct{ DB *db.Queries }

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
		OrderDate   time.Time `json:"order_date"`
		Status      string    `json:"status"`
		ProductID   uuid.UUID `json:"product_id"`
		Price       float64   `json:"price"`
		Qauntity    int       `json:"quantity"`
		TotalAmount float64   `json:"total_amount"`
	}
	var respUpdatedOrderItem respOrderItem
	respUpdatedOrderItem.OrderItemID = updatedOrderItem.ID
	respUpdatedOrderItem.OrderDate = updatedOrderItem.CreatedAt
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
		IsDeleted      bool      `json:"is_deleted"`
		TriggerPrice   float64   `json:"trigger_price"`
		DiscountAmount float64   `json:"discount_amount"`
		DiscountType   string    `json:"discount_type"`
		StartDate      time.Time `json:"start_date"`
		EndDate        time.Time `json:"end_date"`
	}

	var respCoupons []respCoupon
	for _, v := range coupons {
		var temp respCoupon
		temp.ID = v.ID
		temp.Name = v.Name
		temp.IsDeleted = v.IsDeleted
		temp.TriggerPrice = v.TriggerPrice
		temp.DiscountAmount = v.DiscountAmount
		temp.DiscountType = v.DiscountType
		temp.StartDate = v.StartDate
		temp.EndDate = v.EndDate
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
		DiscountType   string  `json:"discount_type"`
		StartDate      string  `json:"start_date"`
		EndDate        string  `json:"end_date"`
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
	if req.DiscountType != utils.CouponDiscountTypeFlat && req.DiscountType != utils.CouponDiscountTypePercentage {
		errors = append(errors, "invalid discount type")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		errors = append(errors, "invalid start date")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		errors = append(errors, "invalid end date")
	}
	if startDate.After(endDate) {
		errors = append(errors, "start date largert than end date")
	}
	// check request start_date and end_date now

	if len(errors) > 0 {
		http.Error(w, strings.Join(errors, "\n"), http.StatusBadRequest)
		return
	}

	coupon, err := a.DB.GetCouponByName(context.TODO(), req.Name)
	if err == sql.ErrNoRows {
	} else if err != nil {
		log.Error("error checking whether coupon already exists")
		http.Error(w, "internal error checking whether coupon already exists", http.StatusInternalServerError)
		return
	} else if coupon.IsDeleted {
		http.Error(w, "trying to create a coupon that already exists and is disabled by admin", http.StatusBadRequest)
		return
	} else if !coupon.IsDeleted {
		http.Error(w, "trying to create a coupon that already exists", http.StatusBadRequest)
	}

	var addCouponArg db.AddCouponParams
	addCouponArg.Name = req.Name
	addCouponArg.TriggerPrice = req.TriggerPrice
	addCouponArg.DiscountAmount = req.DiscountAmount
	addCouponArg.DiscountType = req.DiscountType
	addCouponArg.StartDate = startDate
	addCouponArg.EndDate = endDate
	addedCoupon, err := a.DB.AddCoupon(context.TODO(), addCouponArg)
	if err != nil {
		log.Error("error adding coupon after successful validation:", err.Error())
		http.Error(w, "internal error adding coupon", http.StatusInternalServerError)
		return
	}

	type respCoupon struct {
		Name           string    `json:"name"`
		TriggerPrice   float64   `json:"trigger_price"`
		DiscountAmount float64   `json:"discount_amount"`
		DiscountType   string    `json:"discount_type"`
		StartDate      time.Time `json:"start_date"`
		EndDate        time.Time `json:"end_date"`
	}

	var respCouponData respCoupon
	respCouponData.Name = addedCoupon.Name
	respCouponData.TriggerPrice = addedCoupon.TriggerPrice
	respCouponData.DiscountAmount = addedCoupon.DiscountAmount
	respCouponData.DiscountType = addedCoupon.DiscountType
	respCouponData.StartDate = addedCoupon.StartDate
	respCouponData.EndDate = addedCoupon.EndDate

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
		DiscountType   string  `json:"discount_type"`
		StartDate      string  `json:"start_date"`
		EndDate        string  `json:"end_date"`
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
	if req.DiscountType != utils.CouponDiscountTypeFlat && req.DiscountType != utils.CouponDiscountTypePercentage {
		errors = append(errors, "invalid discount type")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		errors = append(errors, "invalid start date")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		errors = append(errors, "invalid end date")
	}
	if startDate.After(endDate) {
		errors = append(errors, "start date largert than end date")
	}
	// check request start_date and end_date now
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
	editCouponArg.DiscountType = req.DiscountType
	editCouponArg.StartDate = startDate
	editCouponArg.EndDate = endDate

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
		DiscountType   string    `json:"discount_type"`
		StartDate      time.Time `json:"start_date"`
		EndDate        time.Time `json:"end_date"`
		Message        string    `json:"message"`
	}
	data.CouponID = editedCoupon.ID
	data.NewName = editedCoupon.Name
	data.OldName = editCouponArg.OldName
	data.TriggerPrice = editedCoupon.TriggerPrice
	data.DiscountAmount = editedCoupon.DiscountAmount
	data.DiscountType = editedCoupon.DiscountType
	data.StartDate = editedCoupon.StartDate
	data.EndDate = editedCoupon.EndDate
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
	coupon, err := a.DB.GetCouponByName(context.TODO(), couponName)
	if err == sql.ErrNoRows {
		http.Error(w, "coupon does not exist to delete", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Error("error fetchin coupon to check whether coupon is deleted in DeleteCouponHandler for Admin")
		http.Error(w, "internal error fetchin coupon to check whether coupon is deleted", http.StatusInternalServerError)
		return
	} else if coupon.IsDeleted {
		msg := fmt.Sprintf("coupon %s already deleted.", coupon.Name)
		http.Error(w, msg, http.StatusForbidden)
		return
	}
	coupon, err = a.DB.DeleteCouponByName(context.TODO(), couponName)
	if err != nil {
		log.Error("error soft deleting coupon:", err.Error())
		http.Error(w, "internal error soft deleting coupon", http.StatusInternalServerError)
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
	timeLimit := r.URL.Query().Get("time_limit")

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

	dateArg := db.GetVendorPaymentsByDateRangeParams{StartDate: startDate, EndDate: endDate}
	vendorPayments, err := a.DB.GetVendorPaymentsByDateRange(context.TODO(), dateArg)
	if err != nil {
		log.Error("error fetching vendorPayments:", err.Error())
		http.Error(w, "internal error fetching data", http.StatusInternalServerError)
		return
	}

	orderItems, err := a.DB.GetAllOrderItemsForAdmin(context.TODO())
	if err != nil {
		log.Error("error fetching orderItems:", err.Error())
		http.Error(w, "internal error fetching orderItems", http.StatusInternalServerError)
		return
	}

	var sellerOrderAmountMap = make(map[uuid.UUID]float64)
	for _, vp := range vendorPayments {
		sellerOrderAmountMap[vp.SellerID] += vp.TotalAmount
	}

	//Convert to slice
	type sellerStat struct {
		SellerID    uuid.UUID
		SellerName  string
		TotalAmount float64
	}

	var sortedSellers []sellerStat
	for sid, total := range sellerOrderAmountMap {
		seller, err := a.DB.GetUserById(context.TODO(), sid)
		if err != nil {
			log.Println("error fetching seller by sellerID in SalesReportHandler")
			continue
		}
		sortedSellers = append(sortedSellers, sellerStat{
			SellerID:    sid,
			SellerName:  seller.Name,
			TotalAmount: total,
		})
	}

	// Sort descending
	sort.Slice(sortedSellers, func(i, j int) bool {
		return sortedSellers[i].TotalAmount > sortedSellers[j].TotalAmount
	})

	// Take top 3
	topSellers := sortedSellers
	if len(sortedSellers) > 3 {
		topSellers = sortedSellers[:3]
	}

	var (
		totalProfit, totalLossAmount float64
		totalSales, totalOrders      int
		statusCount                  = make(map[string]int)
		orderItemMap                 = make(map[uuid.UUID]map[string]float64)
	)

	totalOrders = len(orderItems)
	for _, oi := range orderItems {
		statusCount[oi.Status]++
	}

	for _, vp := range vendorPayments {
		if vp.Status == utils.StatusVendorPaymentCancelled {
			continue
		}
		totalSales++
		totalProfit += vp.PlatformFee
		if _, exists := orderItemMap[vp.OrderItemID]; !exists {
			orderItemMap[vp.OrderItemID] = map[string]float64{"sales": 0, "platform_fee": 0}
		}
		orderItemMap[vp.OrderItemID]["sales"] += vp.TotalAmount
		orderItemMap[vp.OrderItemID]["platform_fee"] += vp.PlatformFee
	}

	orders, err := a.DB.GetAllOrders(context.TODO())
	if err != nil {
		log.Error("error fetching orders:", err.Error())
		http.Error(w, "internal error fetching orders", http.StatusInternalServerError)
		return
	}

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

	// === Begin PDF ===
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, "Admin Sales Report")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(95, 8, fmt.Sprintf("Start Date: %s", startDate.Format("2006-01-02")))
	pdf.Cell(95, 8, fmt.Sprintf("End Date: %s", endDate.Format("2006-01-02")))
	pdf.Ln(12)

	pdf.SetLineWidth(0.3)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(5)

	// ========== TOP 3 SELLERS BY TOTAL ORDER AMOUNT ========== //
	pdf.SetFont("Arial", "B", 13)
	pdf.Cell(0, 8, "Top 3 Sellers by Total Order Amount")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(230, 230, 230)
	pdf.CellFormat(10, 8, "No.", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Seller Name", "1", 0, "", true, 0, "")
	pdf.CellFormat(90, 8, "Seller ID", "1", 0, "", true, 0, "")
	pdf.CellFormat(30, 8, "Total Amount", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	for i, s := range topSellers {
		pdf.CellFormat(10, 8, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 8, s.SellerName, "1", 0, "", false, 0, "")
		pdf.CellFormat(90, 8, s.SellerID.String(), "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 8, fmt.Sprintf("$%.2f", s.TotalAmount), "1", 1, "R", false, 0, "")
	}
	pdf.Ln(12)

	// Summary Section
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Summary")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(95, 8, fmt.Sprintf("Total Orders: %d", totalOrders))
	pdf.Cell(95, 8, fmt.Sprintf("Total Sales: %d", totalSales))
	pdf.Ln(8)
	pdf.Cell(95, 8, fmt.Sprintf("Pending Orders: %d", statusCount[utils.StatusOrderPending]))
	pdf.Cell(95, 8, fmt.Sprintf("Processing Orders: %d", statusCount[utils.StatusOrderProcessing]))
	pdf.Ln(8)
	pdf.Cell(95, 8, fmt.Sprintf("Shipped Orders: %d", statusCount[utils.StatusOrderShipped]))
	pdf.Cell(95, 8, fmt.Sprintf("Delivered Orders: %d", statusCount[utils.StatusOrderDelivered]))
	pdf.Ln(8)
	pdf.Cell(95, 8, fmt.Sprintf("Cancelled Orders: %d", statusCount[utils.StatusOrderCancelled]))
	pdf.Cell(95, 8, fmt.Sprintf("Returned Orders: %d", statusCount[utils.StatusOrderReturned]))
	pdf.Ln(8)
	pdf.Cell(95, 8, fmt.Sprintf("Platform Profit: $%.2f", totalProfit))
	pdf.Cell(95, 8, fmt.Sprintf("Discount Loss: $%.2f", totalLossAmount))
	pdf.Ln(8)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, fmt.Sprintf("Net Profit: $%.2f", totalProfit-totalLossAmount))
	pdf.Ln(12)

	// Sales Table
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Order Item Sales Breakdown")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(100, 8, "OrderItem ID", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 8, "Total Amount ($)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Platform Fee ($)", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	fill := false
	pdf.SetFillColor(245, 245, 245)
	for id, data := range orderItemMap {
		pdf.CellFormat(100, 8, id.String(), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(50, 8, fmt.Sprintf("%.2f", data["sales"]), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", data["platform_fee"]), "1", 1, "C", fill, 0, "")
		fill = !fill
	}

	// Discounts Table
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Discounted Orders")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(100, 8, "Order ID", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 8, "Discount ($)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Coupon", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	fill = false
	pdf.SetFillColor(245, 245, 245)
	for _, o := range orders {
		pdf.CellFormat(100, 8, o.ID.String(), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(50, 8, fmt.Sprintf("%.2f", o.DiscountAmount), "1", 0, "C", fill, 0, "")
		couponName := "no coupon"
		if o.CouponID.Valid {
			coupon, err := a.DB.GetCouponByID(context.TODO(), o.CouponID.UUID)
			if err == nil {
				couponName = coupon.Name
			}
		}
		pdf.CellFormat(40, 8, couponName, "1", 1, "C", fill, 0, "")
		fill = !fill
	}
	pdf.Ln(10)
	pdf.Ln(10)
	// Charts
	pieChart, err := chartGen.GenerateOrderStatusPieChartForAdmin(
		statusCount[utils.StatusOrderPending],
		statusCount[utils.StatusOrderProcessing],
		statusCount[utils.StatusOrderShipped],
		statusCount[utils.StatusOrderDelivered],
		statusCount[utils.StatusOrderCancelled],
		statusCount[utils.StatusOrderReturned],
	)
	if err == nil {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "Order Status Distribution")
		pdf.Ln(8)
		chartGen.AddChartToPDFForAdmin(pdf, pieChart, "order_status_chart", 15, pdf.GetY(), 90)
		pdf.Ln(75)
	}

	barChart, err := chartGen.GenerateProfitLossBarChartForAdmin(totalProfit, totalLossAmount)
	if err == nil {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "Profit vs Loss Chart")
		pdf.Ln(8)
		chartGen.AddChartToPDFForAdmin(pdf, barChart, "profit_loss_chart", 15, pdf.GetY(), 180)
		pdf.Ln(80)
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
