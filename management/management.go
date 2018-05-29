package management

import (
	"errors"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
)

type MgmtModuleCfg struct {
	DB        *sqlx.DB
	JwtSecret string
	Domain    string
}

type GeneralStore interface {
	Begin() (*sqlx.Tx, error)
	Devices() DeviceStore
	Groups() GroupStore
	Users() UserStore
}

type Module interface {
	GeneralStore
	ApiAuthModule() http.Handler
	ApiGroupModule() http.Handler
}

func NewGeneralStore(cfg MgmtModuleCfg) Module {
	return &mgmtModule{cfg}
}

type mgmtModule struct {
	MgmtModuleCfg
}

func (m *mgmtModule) Begin() (*sqlx.Tx, error) {
	return m.DB.Beginx()
}

func (m *mgmtModule) Devices() DeviceStore {
	return m
}

func (m *mgmtModule) Groups() GroupStore {
	return m
}

func (m *mgmtModule) Users() UserStore {
	return m
}

type IOTDevice struct {
	ID    string
	Name  string
	Mac   string
	IP    string
	Group string
}

func (iot *IOTDevice) GetName() string {
	return iot.Name
}

func (iot *IOTDevice) GetMAC() string {
	return iot.Mac
}

func (iot *IOTDevice) GetIP() string {
	return iot.IP
}

func (iot *IOTDevice) GetGroup() string {
	return iot.Group
}

type Device interface {
	GetName() string
	GetMAC() string
	GetIP() string
	GetGroup() string
}

type DeviceStore interface {
	AddDevice(tx *sqlx.Tx, d Device) error
	GetDeviceByID(tx *sqlx.Tx, id interface{}) (Device, error)
	GetAllByUserID(tx *sqlx.Tx, id interface{}) ([]Device, error)
}

func (m *mgmtModule) AddDevice(tx *sqlx.Tx, d Device) error {
	i := "INSERT INTO DEVICE(id,name,mac,group_id,create_date) VALUES ($1,$2,$3,$4,$5)"
	ID := uuid.NewV4().String()
	_, err := tx.Exec(i, ID, d.GetName(), d.GetMAC(), d.GetGroup(), time.Now())
	return err
}

func (m *mgmtModule) GetDeviceByID(tx *sqlx.Tx, id interface{}) (Device, error) {
	r := tx.QueryRow("SELECT id,name,mac,group_id FROM DEVICE WHERE ID = :id", id)
	d := IOTDevice{}
	if err := r.Scan(&d.ID, &d.Name, &d.Mac, &d.Group); err != nil {
		return nil, err
	}
	return &d, nil
}

func (m *mgmtModule) GetAllByUserID(tx *sqlx.Tx, id interface{}) ([]Device, error) {
	all_by_user := "SELECT id,name,mac,group_id FROM device d join user_group ug using (group_id) where ug.user_id = $1"
	rows, err := tx.Queryx(all_by_user, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lst := []Device{}
	for rows.Next() {
		d := &IOTDevice{}
		err := rows.StructScan(d)
		if err != nil {
			return lst, err
		}
		lst = append(lst, d)
	}
	return lst, nil
}

type GroupStore interface {
	AddGroup(tx *sqlx.Tx, g Group) (Group, error)
	GetGroupsBy(tx *sqlx.Tx, userId string) ([]Group, error)
}

type Group interface {
	GetID() string
	GetOwnerID() string
	GetName() string
}

type IOTGroup struct {
	ID      string
	OwnerID string
	Name    string
}

func (g *IOTGroup) GetID() string {
	return g.ID
}

func (g *IOTGroup) GetOwnerID() string {
	return g.OwnerID
}

func (g *IOTGroup) GetName() string {
	return g.Name
}

type UserStore interface {
	AddUser(*sqlx.Tx, User) error
	GetUser(tx *sqlx.Tx, id string) (User, error)
	GetUserByEmail(tx *sqlx.Tx, email string) (User, error)
	UpdateUser(tx *sqlx.Tx, u User) error
	DeleteUser(tx *sqlx.Tx, id string) (User, error)
}

type User interface {
	GetID() string
	GetName() string
	GetEmail() string
	GetPassword() string
}

func (ds *mgmtModule) AddGroup(tx *sqlx.Tx, g Group) (Group, error) {
	insert := "INSERT INTO DEVICE_GROUP(ID,GROUP_NAME,DESCRIPTION) VALUES ($1,$2,$3)"
	_, err := tx.Exec(insert, g.GetID(), g.GetName(), g.GetOwnerID())
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (ds *mgmtModule) GetGroupsBy(tx *sqlx.Tx, userId string) ([]Group, error) {
	st := "SELECT ID,GROUP_NAME,DESCRIPTION FROM DEVICE_GROUP DG,USER_GROUP_MAPPING DGM WHERE DG.ID = DGM.GROUP_ID AND DGM.USER_ID = $1"
	rows, err := tx.Queryx(st, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lst := []Group{}
	for rows.Next() {
		g := &IOTGroup{}
		err := rows.StructScan(g)
		if err != nil {
			return lst, err
		}
		lst = append(lst, g)
	}
	return lst, nil

}

type AppUser struct {
	ID       string
	Name     string
	Email    string
	Password string
}

func (u *AppUser) GetID() string {
	return u.ID
}

func (u *AppUser) GetName() string {
	return u.Name
}

func (u *AppUser) GetEmail() string {
	return u.Email
}

func (u *AppUser) GetPassword() string {
	return u.Password
}

func (m *mgmtModule) AddUser(tx *sqlx.Tx, u User) error {
	insert := "INSERT INTO app_user(id,name,password,email) values ($1,$2,$3,$4)"
	_, err := tx.Exec(insert, u.GetID(), u.GetName(), u.GetPassword(), u.GetEmail())
	return err
}

func (m *mgmtModule) GetUser(tx *sqlx.Tx, id string) (User, error) {
	sel := "SELECT id,name,email FROM app_user where id = $1"
	r := tx.QueryRow(sel, id)
	u := &AppUser{}
	err := r.Scan(&u.ID, &u.Name, &u.Email)
	return u, err
}

func (m *mgmtModule) GetUserByEmail(tx *sqlx.Tx, mail string) (User, error) {
	sel := "SELECT id,name,email,password FROM app_user where email = $1"
	r := tx.QueryRow(sel, mail)
	u := &AppUser{}
	err := r.Scan(&u.ID, &u.Name, &u.Email, &u.Password)
	return u, err
}

func (m *mgmtModule) UpdateUser(tx *sqlx.Tx, u User) error {
	update := "UPDATE app_user set name= $1, email= $2) where id = $3"
	r, err := tx.Exec(update, u.GetName(), u.GetEmail(), u.GetID())
	if err != nil {
		return err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if affected < 1 {
		return errors.New("ID not found")
	}
	return nil
}

func (m *mgmtModule) DeleteUser(tx *sqlx.Tx, id string) (User, error) {
	panic("not implemented")
}
