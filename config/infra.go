package config

import (
	"fmt"
	"net"
)

type Infra struct {
	MainDatabase      Mongo         `koanf:"mainDatabase"`
	MainCache         Redis         `koanf:"mainCache"`
	SessionsCache     Redis         `koanf:"sessionsCache"`
	MainMessageBroker RabbitMq      `koanf:"mainMessageBroker"`
	MainObjectStorage Minio         `koanf:"mainObjectStorage"`
	MainTextSearch    ElasticSearch `koanf:"mainTextSearch"`
}

type Mongo struct {
	Host     string `koanf:"host" validate:"required,host"`
	Port     int    `koanf:"port" validate:"required,min=1024,max=65535"`
	Username string `koanf:"username" validate:"required,max=32"`
	Password string `koanf:"password" validate:"required,max=32"`
	Name     string `koanf:"name" validate:"required,max=32"`
}

func (m Mongo) Addr() string {
	return net.JoinHostPort(m.Host, fmt.Sprintf("%d", m.Port))
}

type Redis struct {
	Host     string `koanf:"host" validate:"required,host"`
	Port     int    `koanf:"port" validate:"required,min=1024,max=65535"`
	Username string `koanf:"username" validate:"required,max=32"`
	Password string `koanf:"password" validate:"required,max=32"`
	Database int    `koanf:"database" validate:"required,min=0"`
}

func (r Redis) Addr() string {
	return net.JoinHostPort(r.Host, fmt.Sprintf("%d", r.Port))
}

type RabbitMq struct {
	Host     string `koanf:"host" validate:"required,host"`
	Port     int    `koanf:"port" validate:"required,min=1024,max=65535"`
	Username string `koanf:"username" validate:"required,max=32"`
	Password string `koanf:"password" validate:"required,max=32"`
}

func (r RabbitMq) Addr() string {
	return net.JoinHostPort(r.Host, fmt.Sprintf("%d", r.Port))
}

type Minio struct {
	Host     string `koanf:"host" validate:"required,host"`
	Port     int    `koanf:"port" validate:"required,min=1024,max=65535"`
	Username string `koanf:"username" validate:"required,max=32"`
	Passward string `koanf:"password" validate:"required,max=32"`
	UseTLS   bool   `koanf:"useTLS"`
}

func (m Minio) Addr() string {
	return net.JoinHostPort(m.Host, fmt.Sprintf("%d", m.Port))
}

type ElasticSearch struct {
	Host string `koanf:"host" validate:"required,max=32"`
	Port int    `koanf:"port" validate:"required,max=32"`
}

func (es ElasticSearch) Addr() string {
	return net.JoinHostPort(es.Host, fmt.Sprintf("%d", es.Port))
}
