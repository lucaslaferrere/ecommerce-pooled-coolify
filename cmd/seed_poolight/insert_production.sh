#!/usr/bin/env bash
# Run this after Coolify deploys the "allow products without variants" fix.
# Usage: TOKEN=<jwt> bash insert_production.sh

set -e

API="https://api.pooled.com.ar/api/v1/admin/products"
AUTH="Authorization: Bearer $TOKEN"

create() {
  local name="$1" price="$2" desc="$3"
  local result
  result=$(curl -s -X POST "$API" \
    -H "$AUTH" \
    -F "name=$name" \
    -F "description=$desc" \
    -F "base_price=$price" \
    -F "category=luminarias" \
    -F "brand=Pooled" \
    -F "stock=10")
  echo "$name → $result"
}

create "POOLIGHT Wall Mounted Plastic RGBW"  "100575.40" "Luminaria de pared para piscina, cuerpo plástico, luz RGBW."
create "POOLIGHT Wall Mounted Plastic Blanco" "74943.32"  "Luminaria de pared para piscina, cuerpo plástico, luz blanca."
create "POOLIGHT Wall Mounted SS316 RGBW"     "125800.62" "Luminaria de pared para piscina, acero inoxidable SS316, luz RGBW."
create "POOLIGHT Wall Mounted SS316 Blanco"   "92845.09"  "Luminaria de pared para piscina, acero inoxidable SS316, luz blanca."
create "POOLIGHT Mini 1.5\" Plastic RGBW"     "131659.38" "Luminaria mini 1.5\" para piscina, cuerpo plástico, luz RGBW."
create "POOLIGHT Mini 1.5\" Plastic Blanco"   "103179.29" "Luminaria mini 1.5\" para piscina, cuerpo plástico, luz blanca."
create "POOLIGHT Mini 1.5\" SS316 RGBW"       "157494.89" "Luminaria mini 1.5\" para piscina, acero inoxidable SS316, luz RGBW."
create "POOLIGHT Mini 1.5\" SS316 Blanco"     "129014.80" "Luminaria mini 1.5\" para piscina, acero inoxidable SS316, luz blanca."

echo "Done."
