package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/utils"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	log "github.com/sirupsen/logrus"
)

func paymentRoutine() {
	for {
		updateVendorPaymentsAndSellerWallet()
		time.Sleep(3 * time.Hour)
	}
}

func updateVendorPaymentsAndSellerWallet() {
	var dbConn = repository.NewDBConfig()
	var DB = db.New(dbConn)
	orderItems, err := DB.GetAllOrderItemsForAdmin(context.TODO())
	if err != nil {
		log.Error("error fetching orderItems in payment go routine")
		return
	} else if len(orderItems) == 0 {
		log.Info("no current order items")
	}

	for _, oi := range orderItems {
		vp, err := DB.GetVendorPaymentByOrderItemID(context.TODO(), oi.ID)
		if err == sql.ErrNoRows {
			log.Error("no vendor payment associated with the orderItem in payment go routine")
			continue
		} else if err != nil {
			log.Error("error fetching vendor payment for orderItem in payment go routine")
			continue
		}
		if oi.Status == utils.StatusOrderDelivered &&
			oi.UpdatedAt.Add(3*24*time.Hour).Before(time.Now()) &&
			vp.Status != utils.StatusVendorPaymentReceived {
			updatedVP, err := DB.EditVendorPaymentStatusByOrderItemID(context.TODO(), db.EditVendorPaymentStatusByOrderItemIDParams{
				OrderItemID: oi.ID,
				Status:      utils.StatusVendorPaymentReceived,
			})
			if err != nil {
				log.Error("error updating vendor payment in payment go routine")
				continue
			}

			_, err = DB.AddSavingsToWalletByUserID(context.TODO(), db.AddSavingsToWalletByUserIDParams{
				Savings: vp.CreditAmount,
				UserID:  vp.SellerID,
			})
			if err != nil {
				log.Error("error  updating wallet savings in payment go routine")
				continue
			}

			msg := fmt.Sprintf("updated vp: %s from status %s to %s", vp.ID.String(), vp.Status, updatedVP.Status)
			log.Info(msg)
			msg = fmt.Sprintf("credited amount %0.2f from vp %s to seller %s", vp.CreditAmount, vp.ID.String(), vp.SellerID.String())
			log.Info(msg)
		}
	}
}
