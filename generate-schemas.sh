#!/usr/bin/env sh
set -e

host="https://demotenant.dev.mambucloud.com/api/swagger"
targetDir="${TMP_DIR:-"./test"}"

# open https://demotenant.dev.mambucloud.com/apidocs/ google chrome console:
# let options = document.querySelectorAll(".download-url-select option"); [...options].map(o => o.value).join(" ");
schemas="json/clients_v2_swagger.json json/clients_documents_v2_swagger.json json/branches_v2_swagger.json json/centres_v2_swagger.json json/users_v2_swagger.json json/userroles_v2_swagger.json json/comments_v2_swagger.json json/setup_general_v2_swagger.json json/setup_organization_v2_swagger.json json/currencies_v2_swagger.json json/glaccounts_v2_swagger.json json/gljournalentries_v2_swagger.json json/groups_v2_swagger.json json/creditarrangements_v2_swagger.json json/loanproducts_v2_swagger.json json/loans_v2_swagger.json json/loans_transactions_v2_swagger.json json/indexratesources_v2_swagger.json json/depositproducts_v2_swagger.json json/deposits_v2_swagger.json json/deposits_transactions_v2_swagger.json json/bulks_v2_swagger.json json/cards_v2_swagger.json json/fundingsources_v2_swagger.json json/tasks_v2_swagger.json json/database_backup_v2_swagger.json json/communications_messages_v2_swagger.json json/customfieldsets_v2_swagger.json json/customfields_v2_swagger.json json/data_import_v2_swagger.json json/organization_identificationDocumentTemplates_v2_swagger.json json/crons_eod_v2_swagger.json json/organization_transactionChannels_v2_swagger.json json/configuration__organization_v2_swagger.json json/configuration__customfields_v2_swagger.json json/configuration__userroles_v2_swagger.json json/configuration__branches.yaml_v2_swagger.json json/configuration__centres.yaml_v2_swagger.json json/application__status_v2_swagger.json json/accounting_interestaccrual_v2_swagger.json json/apikey_rotation_v2_swagger.json json/installments_v2_swagger.json json/documents_v2_swagger.json json/organization_holidays_v2_swagger.json json/organization_holidays_nonworkingdays_v2_swagger.json json/organization_holidays_general_v2_swagger.json"

for schema in $schemas; do
    curl "$host/$schema" | jq > "$targetDir/$schema"
done