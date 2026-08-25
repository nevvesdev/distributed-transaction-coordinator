package domain

import "time"

type DomainEvent interface {
	NomeEvento() string
	OcorridoEm() time.Time
	IDAgregado() string
}
