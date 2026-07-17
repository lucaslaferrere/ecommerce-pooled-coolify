package services

import (
	"log"
	"net"

	"github.com/oschwald/geoip2-golang"
)

// GeoService resuelve país/ciudad desde una IP usando GeoLite2 (MaxMind).
// Si la base no está disponible, Lookup devuelve strings vacíos y el tracking
// de eventos sigue funcionando (fallback graceful).
type GeoService struct {
	reader *geoip2.Reader
}

// NewGeoService abre la base GeoLite2 desde dbPath. Si dbPath está vacío o no se
// puede abrir, devuelve un GeoService inerte (Lookup → "", "").
func NewGeoService(dbPath string) *GeoService {
	if dbPath == "" {
		log.Println("[geo] GEOIP_DB_PATH no configurado — geolocalización desactivada")
		return &GeoService{}
	}
	reader, err := geoip2.Open(dbPath)
	if err != nil {
		log.Printf("[geo] no se pudo abrir GeoLite2 en %q: %v — geolocalización desactivada", dbPath, err)
		return &GeoService{}
	}
	log.Printf("[geo] GeoLite2 cargada desde %q", dbPath)
	return &GeoService{reader: reader}
}

// Lookup devuelve (country, city, province, lat, lng) para la IP. Valores vacíos/cero
// si la base no está cargada o la IP no se puede resolver.
func (g *GeoService) Lookup(ip string) (country, city, province string, lat, lng float64) {
	if g == nil || g.reader == nil || ip == "" {
		return "", "", "", 0, 0
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", "", "", 0, 0
	}
	rec, err := g.reader.City(parsed)
	if err != nil {
		return "", "", "", 0, 0
	}
	country = firstNonEmpty(rec.Country.Names["es"], rec.Country.Names["en"])
	city = firstNonEmpty(rec.City.Names["es"], rec.City.Names["en"])
	if len(rec.Subdivisions) > 0 {
		province = firstNonEmpty(rec.Subdivisions[0].Names["es"], rec.Subdivisions[0].Names["en"])
	}
	return country, city, province, rec.Location.Latitude, rec.Location.Longitude
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
