package utils

type ContextKey string

const UserKey ContextKey = "user"

const AdminRole = "admin"
const UserRole = "user"
const SellerRole = "seller"

const StatusOrderShipped = "shipped"
const StatusOrderProcessing = "processing"
const StatusOrderPending = "pending"
const StatusOrderDelivered = "delivered"
const StatusOrderCancelled = "cancelled"

const StatusPaymentProcessing = "processing"
const StatusPaymentSuccessful = "successful"
const StatusPaymentFailed = "failed"
const StatusPaymentReturned = "returned"
const StatusPaymentCancelled = "cancelled"

const StatusVendorPaymentWaiting = "waiting"
const StatusVendorPaymentPending = "pending"
const StatusVendorPaymentCancelled = "cancelled"
const StatusVendorPaymentReceived = "received"
const StatusVendorPaymentFailed = "failed"
const PlatformFeePercentage = 0.15

const StatusPaymentMethodCod = "cod"
const StatusPaymentMethodWallet = "wallet"
const StatusPaymentMethodRpay = "razorpay"

const EcomName = "Toy Stores Ecom"
