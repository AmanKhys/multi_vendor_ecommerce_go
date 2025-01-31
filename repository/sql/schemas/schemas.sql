create table if not exists user_addresses (
    id uuid not null primary key default uuid_generate_v4(),
    user_id uuid not null unique,
    house_name text not null,
    street_name text not null,
    town text not null,
    district text not null,
    state text not null,
    pincode integer not null,
    created_at timestamp without time zone not null default current_timestamp,
    updated_at timestamp without time zone not null default current_timestamp,
    constraint updated_at_check check (updated_at >= created_at),
    constraint user_id_fk foreign key (user_id) references users(id)
);

create table if not exists seller_addresses (
    id uuid not null primary key default uuid_generate_v4(),
    seller_id uuid not null unique,
    building_name text not null,
    street_name text not null,
    town text not null,
    district text not null,
    state text not null,
    pincode integer not null,
    created_at timestamp without time zone not null default current_timestamp,
    updated_at timestamp without time zone not null default current_timestamp,
    constraint updated_at_check check (updated_at >= created_at),
    constraint seller_id_fk foreign key (seller_id) references sellers(id)
);

create table if not exists users (
    id uuid not null primary key default uuid_generate_v4(),
    name text not null,
    email text not null unique,
    phone bigint not null,
    password text not null,
    is_blocked boolean not null default false,
    created_at timestamp without time zone not null default current_timestamp,
    updated_at timestamp without time zone not null default current_timestamp,
    constraint email_check check (email ~* '^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$'),
    constraint name_check check (name ~* '^[a-za-z]{3,}[a-za-z ]{0,}$'),
    constraint phone_check check (phone >= 1000000000 and phone <= 9999999999),
    constraint updated_at_check check (updated_at >= created_at)
);

create table if not exists sellers (
    id uuid not null primary key default uuid_generate_v4(),
    name text not null,
    about text not null,
    email text not null unique,
    phone bigint not null unique,
    gst_no text not null unique,
    password text not null,
    is_blocked boolean not null default false,
    created_at timestamp without time zone not null default current_timestamp,
    updated_at timestamp without time zone not null default current_timestamp,
    constraint email_check check (email ~* '^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$'),
    constraint name_check check (name ~* '^[a-za-z]{3,}[a-za-z ]{0,}$'),
    constraint phone_check check (phone >= 1000000000 and phone <= 9999999999),
    constraint updated_at_check check (updated_at >= created_at)
);

create table if not exists categories (
    id uuid not null primary key default uuid_generate_v4(),
    name text not null unique,
    is_deleted boolean not null default false,
    created_at timestamp without time zone not null default current_timestamp,
    updated_at timestamp without time zone not null default current_timestamp,
    constraint name_check check (name ~* '^[a-z0-9]+$'),
    constraint updated_at_check check (updated_at >= created_at)
);

create table if not exists products (
    id uuid not null primary key default uuid_generate_v4(),
    name text not null,
    description text not null,
    price integer not null,
    stock integer not null,
    seller_id uuid not null,
    category_id uuid not null,
    is_deleted boolean not null default false,
    created_at timestamp without time zone not null default current_timestamp,
    updated_at timestamp without time zone not null default current_timestamp,
    constraint name_check check (name ~* '^[a-za-z]{3,}[a-za-z ]{0,}$'),
    constraint price_check check (price > 0),
    constraint stock_check check (stock >= 0),
    constraint updated_at_check check (updated_at >= created_at),
    constraint category_id_fk foreign key (category_id) references categories (id),
    constraint seller_id_fk foreign key (seller_id) references sellers (id)
);

create table if not exists product_images (
    id uuid not null primary key default uuid_generate_v4(),
    product_id uuid not null,
    image_url text not null,
    created_at timestamp without time zone not null default current_timestamp,
    updated_at timestamp without time zone not null default current_timestamp,
    constraint updated_at_check check (updated_at >= created_at),
    constraint product_id_fk foreign key (product_id) references products (id)
);

create table if not exists reviews (
    id uuid not null primary key default uuid_generate_v4(),
    user_id uuid not null,
    product_id uuid not null,
    rating int not null,
    comment text,
    is_deleted boolean not null default false,
    is_edited boolean not null default false,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp,
    constraint rating_check check (rating >= 1 and rating <= 5),
    constraint updated_at_check check (updated_at >= created_at),
    constraint user_id_fk foreign key (user_id) references users (id),
    constraint product_id_fk foreign key (product_id) references products (id)
);

create table if not exists user_login_otps (
    id uuid not null primary key default uuid_generate_v4(),
    user_id uuid not null,
    otp integer not null default floor(random() * 999999),
    created_at timestamp without time zone not null default current_timestamp,
    expires_at timestamp without time zone not null default (current_timestamp + interval '10 minutes'),
    constraint user_id_fk foreign key (user_id) references users(id)
);

create table if not exists seller_login_otps (
    id uuid not null primary key default uuid_generate_v4(),
    seller_id uuid not null,
    otp integer not null default floor(random() * 999999),
    created_at timestamp without time zone not null default current_timestamp,
    expires_at timestamp without time zone not null default (current_timestamp + interval '10 minutes'),
    constraint seller_id_fk foreign key (seller_id) references sellers(id)
);

