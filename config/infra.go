package config

import (
	"fmt"
	"net"
)

type Infra struct {
	MainDatabase      Mongo         `koanf:"mainDatabase"`
	MainCache         Redis         `koanf:"mainCache"`
	SessionsCache     Redis         `koanf:"sessionsCache"`
	MainObjectStorage Minio         `koanf:"mainObjectStorage"`
	MainTextSearch    ElasticSearch `koanf:"mainTextSearch"`
}

type Mongo struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	Name     string `koanf:"name"`
}

func (m Mongo) Addr() string {
	return net.JoinHostPort(m.Host, fmt.Sprintf("%d", m.Port))
}

type Redis struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	Database int    `koanf:"database"`
}

func (r Redis) Addr() string {
	return net.JoinHostPort(r.Host, fmt.Sprintf("%d", r.Port))
}

type Minio struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Username string `koanf:"username"`
	Passward string `koanf:"password"`
	UseTLS   bool   `koanf:"useTLS"`
}

func (m Minio) Addr() string {
	return net.JoinHostPort(m.Host, fmt.Sprintf("%d", m.Port))
}

type ElasticSearch struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

func (es ElasticSearch) Addr() string {
	return net.JoinHostPort(es.Host, fmt.Sprintf("%d", es.Port))
}
