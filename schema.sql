
drop if exists ${SCHEMA_NAME}.device_group;
drop if exists ${SCHEMA_NAME}.user_group_mapping;
drop if exists ${SCHEMA_NAME}.device;

-- device groups
create table ${SCHEMA_NAME}.device_group {
    id text not null,
    group_name text,
    owner_id text,
    description text,
    CONSTRAINT PK_DEVICE_GROUP PRIMARY KEY (id)
};

-- user to group mapping
create table ${SCHEMA_NAME}.user_group_mapping {
    id text not null,
    user_id text,
    group_id text,
    CONSTRAINT PK_USER_GROUP_MAPPING PRIMARY KEY (id)
};

-- device table
create table ${SCHEMA_NAME}.device {
    id text not null,
    mac text,
    name text,
    ip text,
    group_id text not null,
    CONSTRAINT PK_DEVICE PRIMARY KEY (id)
};

create table ${SCHEMA_NAME}.user (
id text not null,
name text,
password text,
email text
);


