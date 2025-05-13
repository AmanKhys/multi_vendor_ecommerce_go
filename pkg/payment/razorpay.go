package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/envname"
	rp "github.com/razorpay/razorpay-go"
)

func ExecuteRazorpay(orderPrice float64) (string, error) {
	rpID := os.Getenv(envname.RPID)
	rpSecretKey := os.Getenv(envname.RPSecretKey)

	client := rp.NewClient(rpID, rpSecretKey)
	data := map[string]any{
		"amount":   int(orderPrice) * 100,
		"currency": "INR",
		"receipt":  "101",
	}

	body, err := client.Order.Create(data, nil)
	if err != nil {
		return "", errors.New("payment not initiated")
	}
	rpOrderID, _ := body["id"].(string)
	return rpOrderID, nil
}

// verify razorpay payment
func VerifyRazorpaySignature(orderID, paymentID, signature string) bool {
	secret := os.Getenv(envname.RPSecretKey)

	// Create a signature from order_id and payment_id using HMAC SHA256
	message := orderID + "|" + paymentID
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return expectedSignature == signature
}
