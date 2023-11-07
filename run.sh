#!/bin/bash

curl "https://${ONCALL_URL}/api/v1/integrations/" \
	--request POST \
	--header "Authorization: ${ONCALL_TOKEN}" \
	--header "Content-Type: application/json" \
	--data '{
      "type":"webhook",
      "templates": {
        "web" : {}
      }
  }'

# curl -vvv "https://${ONCALL_URL}/api/v1/integrations/abcacacas/" \
# 	--request GET \
# 	--header "Authorization: ${ONCALL_TOKEN}" \
# 	--header "Content-Type: application/json"
