package main

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const PCP_TPL = `/usr/share/pgpool2/pcp.conf.template`
const PCP_CFG = `/etc/pgpool2/pcp.conf`
const PGP_CFG = `/etc/pgpool2/pgpool.conf`
const PGP_TPL = `/usr/share/pgpool2/pgpool.conf.template`

type PcpConf struct {
	Users []PcpUser
}
type PcpUser struct {
	User     string
	Password string
}

func (c *PcpConf) parseEnv() {
	value := strings.Split(env("PCP_USERS", "postgres:postgres"), ",")
	c.Users = make([]PcpUser, 0, len(value))
	for _, v := range value {
		val := strings.Split(v, ":")
		if len(val) != 2 {
			panic(fmt.Sprintf("invalid PCP user %s", v))
		}
		c.Users = append(c.Users, PcpUser{
			User:     val[0],
			Password: md5Hex(val[1]),
		})
	}
}
func (c *PcpConf) Generate() {
	if pgc, er := os.OpenFile(PCP_CFG, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600); er != nil {
		panic(er)
	} else {
		tpl, er := template.ParseFiles(PCP_TPL)
		if er != nil {
			panic(er)
		}
		c.parseEnv()
		if er := tpl.Execute(pgc, c); er != nil {
			panic(er)
		}
		pgc.Close()
	}
}

type PgpoolConf struct {
	ListenAddresses       string
	Port                  int
	PcpListenAddresses    string
	PcpPort               int
	Backends              []Backend
	PoolPasswd            string
	AuthenticationTimeout int
	//pool
	NumInitChildren int
	MaxPool         int
	//life time
	ChildLifeTime       int
	ChildMaxConnections int
	ConnectionLifeTime  int
	ClientIdleLimit     int
	//replication
	ReplicationMode                      bool
	replicate_select                     bool
	insert_lock                          bool
	lobj_lock_table                      string
	replication_stop_on_mismatch         bool
	failover_if_affected_tuples_mismatch bool
	//load_balance
	LoadBalanceMode                   bool
	ignore_leading_white_space        bool
	white_function_list               string
	black_function_list               string
	database_redirect_preference_list string
	app_name_redirect_preference_list string
	allow_sql_comments                bool
	//master_slave_mode
	MasterSlaveMode       bool
	master_slave_sub_mode string //stream, slony, logical
	sr_check_period       int    //0
	sr_check_user         int    //'nobody'
	sr_check_password     string //''
	sr_check_database     string //'postgres'
	delay_threshold       int    //0
}

func (c *PgpoolConf) Generate() {
	if pgc, er := os.OpenFile(PGP_CFG, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600); er != nil {
		panic(er)
	} else {
		tpl, er := template.ParseFiles(PGP_TPL)
		if er != nil {
			panic(er)
		}
		c.parseEnv()
		if er := tpl.Execute(pgc, c); er != nil {
			panic(er)
		}
	}
}
func (c *PgpoolConf) parseEnv() {
	c.ListenAddresses = env("PGPOOL_LISTEN_ADDRESSES", "*")
	c.PcpListenAddresses = env("PCP_LISTEN_ADDRESSES", "*")
	c.PoolPasswd = env("POOL_PASSWD", "PoolPasswd")
	c.Port = envi("PGPOOL_LISTEN_ADDRESSES", 5432)
	c.AuthenticationTimeout = envi("PGPOOL_LISTEN_ADDRESSES", 5432)
	c.PcpPort = envi("TCP_LISTEN_ADDRESSES", 9898)
	val := strings.Split(env("PGPOOL_BACKENDS", "1:localhost:5432"), ",")
	c.Backends = make([]Backend, 0, len(val))
	for _, v := range val {
		l := strings.Split(v, ":")
		if len(l) != 3 && len(l) != 4 && len(l) != 5 && len(l) != 6 {
			panic(fmt.Sprintf("Invalid Backend: %s", l))
		}
		c.Backends = append(c.Backends, Backend{
			Order:         forceInt(l[0]),
			Host:          l[1],
			Port:          forceInt(l[2]),
			DataDirectory: ifLen(l, 4, "/var/lib/pgsql/data"),
			Weight:        ifLenI(l, 3, 1),
			Flag:          ifLen(l, 4, "ALLOW_TO_FAILOVER"),
		})
	}
	c.AuthenticationTimeout = envi("PGPOOL_AUTHENTICATION_TIMEOUT", 60)
	c.NumInitChildren = envi("PGPOOL_NUM_INIT_CHILDREN", 32)
	c.MaxPool = envi("PGPOOL_MAX_POOL", 4)
	c.ChildLifeTime = envi("PGPOOL_CHILD_LIFE_TIME", 300)
	c.ChildMaxConnections = envi("PGPOOL_CHILD_MAX_CONNECTIONS", 0)
	c.ConnectionLifeTime = envi("PGPOOL_CONNECTION_LIFE_TIME", 0)
	c.ClientIdleLimit = envi("PGPOOL_CLIENT_IDLE_LIMIT", 0)
	c.ReplicationMode = envb("PGPOOL_REPLICATION_MODE", false)
	c.LoadBalanceMode = envb("PGPOOL_LOAD_BALANCE_MODE", false)
	c.MasterSlaveMode = envb("PGPOOL_MASTER_SLAVE_MODE", false)
}

type Backend struct {
	Order         int
	Host          string
	Port          int
	DataDirectory string
	Weight        int
	Flag          string //DISALLOW_TO_FAILOVER|ALLOW_TO_FAILOVER|ALWAYS_MASTER
}

func main() {
	PGP := env("PGPOOL_CONF_FILE", PGP_CFG)
	PCP := env("PCP_CONF_FILE", PCP_CFG)
	if PCP == PCP_CFG {
		pc := new(PcpConf)
		pc.Generate()
	}
	if PGP == PGP_CFG {
		pp := new(PgpoolConf)
		pp.Generate()
	}
	log.Println("start pgpool")
	cmd := exec.Command("pgpool", "-n", "-f", PGP, "-F", PCP)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if er := cmd.Start(); er != nil {
		log.Panic(er)
	}
	if er := cmd.Wait(); er != nil {
		log.Panic(er)
	}
}

func env(key string, def string) string {
	env := os.Getenv(key)
	if len(env) == 0 {
		env = def
	}
	return env
}
func envb(key string, def bool) bool {
	env := os.Getenv(key)
	if len(env) == 0 {
		return def
	}
	return true
}
func envi(key string, def int) int {
	env := os.Getenv(key)
	if len(env) == 0 {
		return def
	}
	if port, er := strconv.Atoi(env); er != nil {
		return def
	} else {
		return port
	}
}
func forceInt(v string) int {
	if v, e := strconv.Atoi(v); e != nil {
		panic(e)
	} else {
		return v
	}
}
func ifLen(v []string, idx int, def string) string {
	if len(v) >= idx+1 {
		return v[idx]
	} else {
		return def
	}
}
func ifLenI(v []string, idx int, def int) int {
	val := ifLen(v, idx, strconv.Itoa(def))
	if v, e := strconv.Atoi(val); e != nil {
		return def
	} else {
		return v
	}
}
func md5Hex(v string) string {
	return fmt.Sprintf("%x", md5.New().Sum([]byte(v)))
}
