package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func OrdersCron() {
	for {
		cancelVoidOrders()
		time.Sleep(10 * time.Minute)
	}
}

func cancelVoidOrders() {
	var dbConn = repository.NewDBConfig()
	var DB = db.New(dbConn)
	orders, err := DB.GetAllOrders(context.TODO())
	if err != nil {
		log.Error("error fetching orders inn CancelVoidOrders:", err.Error())
		return
	}
	for _, o := range orders {
		payment, err := DB.GetPaymentByOrderID(context.TODO(), o.ID)
		if err == sql.ErrNoRows {
			log.Error("no payment related to the order:", o.ID.String())
			log.Error(payment)
			continue
		} else if err != nil {
			log.Error("error fetching payment for order in CancelVoidOrders:", err.Error())
			continue
		}
		if payment.Method == utils.StatusPaymentMethodCod || payment.Method == utils.StatusPaymentMethodWallet {
			continue
		}

		// otherwise it is razorpay.. check whether paid
		// within time limit of 10 minutes; otherwise cancel the order
		// and cancel the payment, vendor_payments
		fmt.Println(time.Since(o.CreatedAt), o.CreatedAt)
		if (payment.Status == utils.StatusPaymentFailed ||
			payment.Status == utils.StatusPaymentProcessing) &&
			time.Since(o.CreatedAt) > time.Minute*10 {
			orderItems, err := DB.CancelOrderByID(context.TODO(), o.ID)
			if err != nil {
				log.Error("error cancelling order in CancelVoidOrders:", err.Error())
			}

			payment, err := DB.CancelPaymentByOrderID(context.TODO(), o.ID)
			if err != nil {
				log.Error("error cancelling payment in CancelVoidOrders:", err.Error())
			} else {
				log.Info("cancelled payment for order:", o.ID.String())
			}

			vendorPayments, err := DB.CancelVendorPaymentsByOrderID(context.TODO(), o.ID)
			if err != nil {
				log.Error("error cancelling vendor payments  in CancelVoidOrders:", err.Error())
			}

			type PrintOrderItem struct {
				OrderItemID uuid.UUID `json:"order_item_id"`
				Status      string    `json:"order_item_status"`
			}
			type PrintPayment struct {
				PaymentID uuid.UUID `json:"payment_id"`
				Status    string    `json:"payment_status"`
			}
			type PrintVendorPayment struct {
				VendorPaymentID uuid.UUID `json:"vendor_payment_id"`
				Status          string    `json:"vendor_payment_status"`
			}
			var PrintLog struct {
				OrderID        uuid.UUID            `json:"order_id"`
				Payment        PrintPayment         `json:"payment"`
				OrderItems     []PrintOrderItem     `json:"order_items"`
				VendorPayments []PrintVendorPayment `json:"vendor_payments"`
			}

			PrintLog.OrderID = o.ID
			var tempPrintPayment PrintPayment
			tempPrintPayment.PaymentID = payment.ID
			tempPrintPayment.Status = payment.Status
			PrintLog.Payment = tempPrintPayment
			for _, oi := range orderItems {
				var temp PrintOrderItem
				temp.OrderItemID = oi.ID
				temp.Status = oi.Status
				PrintLog.OrderItems = append(PrintLog.OrderItems, temp)
			}

			for _, vp := range vendorPayments {
				var temp PrintVendorPayment
				temp.VendorPaymentID = vp.ID
				temp.Status = vp.Status
				PrintLog.VendorPayments = append(PrintLog.VendorPayments, temp)
			}

			prettyLog := log.New()
			prettyLog.SetFormatter(&log.JSONFormatter{PrettyPrint: true})
			prettyLog.Info(PrintLog)
		}
	}
}
