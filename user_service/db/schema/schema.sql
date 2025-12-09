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

-- User/Seller wallet table
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    savings NUMERIC(10,2) NOT NULL CHECK (savings>=0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP CHECK (updated_at>= created_at)
);
