all: schema_global_ghp_recipient.json \
  schema_global_ghp_vaccination_credential.json \
  schema_global_ghp_covid19_recovery_credential.json \
  schema_global_ghp_covid19_antigen_test_credential.json \
  recipient_overlays \
	vaccination_credential_overlays \
	covid19_recovery_credential_overlays \
	covid19_antigen_test_credential_overlays


schema_global_ghp_recipient.json: spec_global_recipient_schema.json
	layers import csv --spec $< --layerId https://lfph.io/ghp/recipient/schema GHP_data_capture_global.csv > $@

schema_global_ghp_vaccination_credential.json: spec_global_vaccination_credential_schema.json
	layers import csv --spec $< --layerId https://lfph.io/ghp/vaccinationCredential/schema GHP_data_capture_global.csv > $@

schema_global_ghp_covid19_recovery_credential.json: spec_global_covid19_recovery_credential_schema.json 
	layers import csv --spec $< --layerId https://lfph.io/ghp/COVID19RecoveryCredential/schema GHP_data_capture_global.csv > $@

schema_global_ghp_covid19_antigen_test_credential.json: spec_global_covid19_antigen_test_credential_schema.json 
	layers import csv --spec $< --layerId https://lfph.io/ghp/COVID19AntigenTestCredential/schema GHP_data_capture_global.csv > $@

.PHONY: recipient_overlays
recipient_overlays:
	for f in de en-ca en-gb en-us es-us es fi fr-ca fr it pl ; do \
	layers import csv --spec spec_global_recipient_overlay.json --layerId https://lfph.io/ghp/recipient/ovl/$$f GHP_data_capture_global-$$f.csv > overlay_global_ghp_recipient_$$f.json; \
	done

.PHONY: vaccination_credential_overlays
vaccination_credential_overlays:
	for f in de en-ca en-gb en-us es-us es fi fr-ca fr it pl ; do \
	layers import csv --spec spec_global_vaccination_credential_overlay.json --layerId https://lfph.io/ghp/vaccinationCredential/ovl/$$f GHP_data_capture_global-$$f.csv > overlay_global_ghp_vaccination_credential_$$f.json; \
	done

.PHONY: covid19_recovery_credential_overlays
covid19_recovery_credential_overlays:
	for f in de en-ca en-gb en-us es-us es fi fr-ca fr it pl ; do \
	layers import csv --spec spec_global_covid19_recovery_credential_overlay.json --layerId https://lfph.io/ghp/COVID19RecoveryCredential/ovl/$$f GHP_data_capture_global-$$f.csv > overlay_global_ghp_covid19_recovery_credential_$$f.json; \
	done

.PHONY: covid19_antigen_test_credential_overlays
covid19_antigen_test_credential_overlays:
	for f in de en-ca en-gb en-us es-us es fi fr-ca fr it pl ; do \
	layers import csv --spec spec_global_covid19_antigen_test_credential_overlay.json --layerId https://lfph.io/ghp/COVID19AntigenTestCredential/ovl/$$f GHP_data_capture_global-$$f.csv > overlay_global_ghp_covid19_antigen_test_credential_$$f.json; \
	done

