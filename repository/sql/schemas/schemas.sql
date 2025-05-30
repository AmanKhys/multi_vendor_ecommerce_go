-- set timezone of the postgres database to indian time
SET TIMEZONE = 'Asia/Kolkata';
-- Users Table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL CHECK (name ~* '^[a-zA-Z]{3,}[a-zA-Z ]*$'),
    email TEXT NOT NULL UNIQUE CHECK (email ~* '^(0[1-9]|1[0-9]|2[0-9]|3[0-5])[A-Z]{5}[0-9]{4}[A-Z][1-9A-Z]Z[0-9A-Z]$'),
    phone BIGINT UNIQUE CHECK (phone >= 1000000000 AND phone <= 9999999999),
    password TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('user', 'seller', 'admin')),
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    user_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    gst_no TEXT UNIQUE,
    about TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (updated_at >= created_at)
);

-- Addresses Table
CREATE TABLE IF NOT EXISTS addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('user', 'seller')),
    building_name TEXT NOT NULL,
    street_name TEXT NOT NULL,
    town TEXT NOT NULL,
    district TEXT NOT NULL,
    state TEXT NOT NULL,
    pincode INTEGER NOT NULL CHECK (pincode >= 100000 AND pincode <= 999999),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (updated_at >= created_at)
);

-- add partial index after creating address schema 
-- for unique userID when type = 'seller'
CREATE UNIQUE INDEX unique_seller_address_per_user 
ON addresses(user_id) 
WHERE type = 'seller';

-- Categories Table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE CHECK (name ~* '^[a-z0-9]+[a-z0-9 ]*$'),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (updated_at >= created_at)
);

-- Category Items Table
CREATE TABLE IF NOT EXISTS category_items (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id uuid NOT NULL REFERENCES products(id),
    category_id uuid NOT NULL REFERENCES categories(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK(updated_at >= created_at),
    CONSTRAINT category_items_product_category_unique UNIQUE (product_id, category_id)
);

-- Products Table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL CHECK (name ~* '^[a-zA-Z0-9]{3,}[a-zA-Z0-9 ]*$'),
    description TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL CHECK (price > 0),
    stock INTEGER NOT NULL CHECK (stock >= 0),
    seller_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (updated_at >= created_at)
);

-- Product Images Table
CREATE TABLE IF NOT EXISTS product_images (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (updated_at >= created_at)
);

-- Reviews Table
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    is_edited BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (updated_at >= created_at)
);

-- Login OTPs Table
CREATE TABLE IF NOT EXISTS otps(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    otp INTEGER NOT NULL DEFAULT FLOOR(RANDOM() * 999999),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '10 minutes')
);

-- Forgot Password OTPs Table
CREATE TABLE IF NOT EXISTS forgot_otps(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    otp INTEGER NOT NULL DEFAULT FLOOR(RANDOM() * 999999),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '10 minutes')
);

-- Sessions Table
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '7 days')
);

-- Wishlists Table
CREATE TABLE IF NOT EXISTS wishlists (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT cart_user_id_product_id_unique UNIQUE(user_id, product_id)
);

-- Carts Table
CREATE TABLE IF NOT EXISTS carts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INT NOT NULL CHECK (quantity >0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (updated_at>=created_at),
    CONSTRAINT cart_user_id_product_id_unique UNIQUE(user_id, product_id)
);

-- User/Seller wallet table
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    savings NUMERIC(10,2) NOT NULL CHECK (savings>=0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (updated_at>= created_at)
);

-- user orders table
CREATE TABLE orders (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total_amount NUMERIC(10,2) NOT NULL DEFAULT -1, -- -1 when order is first created, later will be updated
    coupon_id UUID REFERENCES coupons(id),
    discount_amount NUMERIC(10,2) DEFAULT 0,
    net_amount NUMERIC(10,2) GENERATED ALWAYS AS (total_amount - discount_amount) STORED,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (updated_at >= created_at)
);


CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    method TEXT NOT NULL CHECK (method in ('razorpay', 'cod', 'wallet')),
    status TEXT NOT NULL CHECK (status in ('not paid', 'processing', 'successful', 'failed', 'cancelled')),
    total_amount NUMERIC(10,2) NOT NULL CHECK(total_amount>0),
    transaction_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK( updated_at>= created_at)
);

CREATE TABLE IF NOT EXISTS shipping_address (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    house_name TEXT NOT NULL,
    street_name TEXT NOT NULL,
    town TEXT NOT NULL,
    district TEXT NOT NULL,
    state TEXT NOT NULL,
    pincode INTEGER NOT NULL CHECK (pincode >= 100000 AND pincode <= 999999),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (updated_at >= created_at)
);

CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    price NUMERIC(10,2) NOT NULL CHECK(price>0),
    quantity INT NOT NULL CHECK (quantity>0),
    -- check == 0 since the orderItems cannot have 0 for total_amount 
    -- thus total_amount here never becomes zero  unless all the items are cancelled.
    total_amount NUMERIC(10, 2) NOT NULL CHECK (total_amount >=0),
    status TEXT NOT NULL CHECK (status in ('pending', 'processing', 'shipped', 'delivered', 'cancelled', 'returned')) DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK(updated_at>=created_at)
);


CREATE TABLE IF NOT EXISTS vendor_payments (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    order_item_id UUID NOT NULL REFERENCES order_items(id),
    seller_id UUID NOT NULL REFERENCES users(id),
    status TEXT NOT NULL CHECK (status in ('waiting','cancelled', 'pending', 'received', 'failed')) default 'waiting',
    total_amount NUMERIC(10, 2) NOT NULL,
    platform_fee NUMERIC(10,2) NOT NULL,
    credit_amount NUMERIC(10,2) NOT NULL ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK(updated_at>=created_at)
);


CREATE TABLE IF NOT EXISTS coupons (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(), 
    name TEXT UNIQUE NOT NULL CHECK (name ~ '^[A-Z0-9]{3,}$'),
    discount_type TEXT NOT NULL CHECK (discount_type in ('flat', 'percentage')),
    trigger_price NUMERIC(10, 2) NOT NULL CHECK (trigger_price > 0),
    discount_amount NUMERIC(10, 2) NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    start_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (start_date<= end_date),
    CHECK (
        (discount_type = 'flat' AND discount_amount <= trigger_price) OR
        (discount_type = 'percentage' AND discount_amount >= 1 AND discount_amount <= 99)
    )
);

CREATE TABLE IF NOT EXISTS return_refunds (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    order_item_id UUID NOT NULL REFERENCES order_items(id),
    payment_id UUID NOT NULL REFERENCES payments(id),
    item_amount NUMERIC(10,2) NOT NULL, -- item amount from order_items[id][total_amount],
    discount_removal_amount NUMERIC(10,2) NOT NULL DEFAULT 0, -- if coupon no longer applicable add the discount amount here
    refund_amount NUMERIC(10,2) NOT NULL GENERATED ALWAYS AS (item_amount - discount_removal_amount) STORED,
    status TEXT NOT NULL CHECK (status in ('refunded', 'not refunded')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (updated_at>=created_at)
);
