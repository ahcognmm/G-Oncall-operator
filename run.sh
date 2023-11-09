#!/bin/bash

# curl "https://${ONCALL_URL}/api/v1/integrations/" \
# 	--request POST \
# 	--header "Authorization: ${ONCALL_TOKEN}" \
# 	--header "Content-Type: application/json" \
# 	--data '{
#       "type":"webhook",
#       "templates": {
#         "web" : {}
#       }
#   }'
#
# curl -vvv "https://${ONCALL_URL}/api/v1/integrations/abcacacas/" \
# 	--request GET \
# 	--header "Authorization: ${ONCALL_TOKEN}" \
# 	--header "Content-Type: application/json"
#
curl "${ONCALL_URL}/api/v1/escalation_policies/ENDWDL4MV8Q2K/" \
	--request PUT \
	--header "Authorization: ${ONCALL_TOKEN}" \
	--header "Content-Type: application/json" \
	--data '{
    "position" : 0,
    "id" : "ENDWDL4MV8Q2K",
    "escalation_chain_id" : "F7SX4Z3VPABJN",
    "type" : "wait",
    "duration": 60
  }'
