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
