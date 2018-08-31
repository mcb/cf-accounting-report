#! /usr/bin/env bash
#
#  Get the reports from the app_usage app and send them through email.
#
set -eu

CF_BASE_DOMAIN=""
USER="admin"
PASSWORD=""
PCFENV=""

cf login --skip-ssl-validation -u $USER -p $PASSWORD -a https://api.$CF_BASE_DOMAIN -o system -s system

echo -e "\nGetting app usage information...\n"
curl "https://app-usage.${CF_BASE_DOMAIN}/system_report/app_usages" -k --silent -H "authorization: `cf oauth-token`" > out/app_usages.json

echo -e "\nGetting service usage information...\n"
curl "https://app-usage.${CF_BASE_DOMAIN}/system_report/service_usages" -k --silent -H "authorization: `cf oauth-token`" > out/service_usages.json

echo -e "\n\nLong term overview app usage:\n"
cat out/app_usages.json | jq -r '.monthly_reports[] | (.year|tostring) + "-" + (.month|tostring) + " : " + (.average_app_instances|tostring)' | cut -d. -f1

echo -e "\napp_usages:\n"
cat out/app_usages.json | jq .

echo -e "\nservice_usages:\n"
cat out/service_usages.json | jq .

# combine the two in an array:
cat out/app_usages.json out/service_usages.json | jq -s . > out/app_service_usages.json
