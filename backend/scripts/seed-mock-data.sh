#!/usr/bin/env bash
# Seeds a local backend with a few Swedish mock recipes.
# Usage: backend/scripts/seed-mock-data.sh [API_BASE_URL]   (default http://localhost:8080)
# Env:   ADMIN_PIN (default 123456, the development default). Writes require an admin token.
set -euo pipefail

BASE="${1:-http://localhost:8080}/api/v1"
API="$BASE/recipes"
PIN="${ADMIN_PIN:-123456}"

TOKEN=$(curl -fsS -X POST "$BASE/auth/login" -H 'Content-Type: application/json' \
  -d "{\"pin\":\"$PIN\"}" | sed -E 's/.*"token":"([^"]*)".*/\1/')

post() {
  curl -fsS -X POST "$API" -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d "$1" > /dev/null
  echo "Skapade: $(echo "$1" | sed -E 's/.*"title":"([^"]*)".*/\1/')"
}

post '{"title":"Pannkakor","description":"Klassiska tunna svenska pannkakor.","categories":["breakfast"],"ingredients":["3 dl vetemjöl","6 dl mjölk","3 ägg","1 krm salt","smör till stekning"],"steps":["Vispa ihop mjöl och hälften av mjölken till en slät smet.","Vispa i resten av mjölken, äggen och saltet.","Stek tunna pannkakor i smör i en het panna."]}'

post '{"title":"Köttbullar","description":"Husmanskost med gräddsås och lingonsylt.","categories":["dinner"],"ingredients":["500 g blandfärs","1 dl ströbröd","1 dl mjölk","1 ägg","1 gul lök","1 tsk salt","smör till stekning"],"steps":["Blanda ströbröd och mjölk och låt svälla.","Rör ihop alla ingredienser till en smidig smet.","Rulla bollar och stek dem i smör tills de är genomstekta."],"oven_temperature":175}'

post '{"title":"Kladdkaka","description":"Seg och chokladig kaka som serveras med grädde.","categories":["dessert","baking"],"ingredients":["100 g smör","2 ägg","2 dl socker","1,5 dl vetemjöl","4 msk kakao","1 tsk vaniljsocker","1 krm salt"],"steps":["Sätt ugnen på 175 grader och smöra en form.","Smält smöret och rör ner övriga ingredienser.","Grädda i cirka 15 minuter så att kakan förblir kladdig."],"oven_temperature":175}'

post '{"title":"Kycklinggryta","description":"Snabb vardagsgryta med curry och ris.","categories":["dinner"],"ingredients":["600 g kycklingfilé","1 gul lök","2 msk currypasta","4 dl matlagningsgrädde","ris att servera till"],"steps":["Skär kycklingen i bitar och fräs tillsammans med löken.","Rör i currypastan och tillsätt grädden.","Låt sjuda i 15 minuter och servera med ris."]}'

post '{"title":"Filmjölk med müsli","description":"Enkel frukost på två minuter.","categories":["breakfast"],"ingredients":["3 dl filmjölk","1 dl müsli","1 banan"],"steps":["Häll filmjölk i en skål.","Toppa med müsli och skivad banan."]}'
