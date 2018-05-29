package management

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const TEST_CONNSTRING = "postgres://pgtest:pgtest@localhost:5432/test_database?sslmode=disable"
const TEST_DRIVE = "postgres"

var test_cfg MgmtModuleCfg

var db *sqlx.DB

func beforeModule(db *sqlx.DB) {
	db.MustExec(create_ddl)
	db.MustExec(test_data)
}

func tearDownModule(db *sqlx.DB) {
	db.Exec(drop_ddl)
}

type ModuleTestFunc func(m GeneralStore)

func testWithModule(tf ModuleTestFunc) {
	module, err := NewGeneralStore(test_cfg)
	if err != nil {
		fmt.Println(err)
	}
	tf(module)
}

func TestMain(m *testing.M) {
	db, err := sqlx.Connect(TEST_DRIVE, TEST_CONNSTRING)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
	test_cfg = MgmtModuleCfg{db}
	defer db.Close()
	beforeModule(db)
	m.Run()
	tearDownModule(db)
}

func Test_mgmtModule_AddDevice(t *testing.T) {
	testWithModule(func(module GeneralStore) {
		tx, err := module.Begin()
		if err != nil {
			fmt.Println(err)
			t.FailNow()
		}
		d := IOTDevice{"test_device_1", "test", "MAC", "1234", "group_1"}
		err = module.Devices().AddDevice(tx, &d)
		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}
		tx.Commit()
	})
}

func Test_mgmtModule_AddGroup(t *testing.T) {
	testWithModule(func(module GeneralStore) {
		tx, err := module.Begin()
		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}
		module.Groups().AddGroup(tx, &IOTGroup{"1234", "Name", "Grupicka"})
		tx.Commit()
	})
}

const drop_ddl = `
drop table if exists device_group;
drop table if exists user_group_mapping;
drop table if exists device;
drop table if exists user;
`

const create_ddl = `
drop table if exists device_group;
drop table if exists user_group_mapping;
drop table if exists device;
drop table if exists app_user;

-- device groups
create table device_group (
    id text,
	group_name text,
	owner_id text,
    description text,
    CONSTRAINT PK_DEVICE_GROUP PRIMARY KEY (id)
);

-- user to group mapping
create table user_group_mapping (
    id text,
    user_id text,
    group_id text,
    CONSTRAINT PK_USER_GROUP_MAPPING PRIMARY KEY (id)
);

-- device table
create table device (
    id text,
    mac text,
    name text,
    ip text,
	group_id text not null,
	create_date timestamp,
    CONSTRAINT PK_DEVICE PRIMARY KEY (id)
);

-- user table
create table app_user (
	id text not null,
	name text not null,
	password text not null,
	email text
);
`

const test_data = `
insert into app_user(id,name,password,email) values ('user_1','user one','password','user_1@test.com');
insert into device_group(id,group_name) values ('group_1','group one');
insert into user_group_mapping(id,user_id,group_id) values ('mapping_1','user_1','group_1');
`
