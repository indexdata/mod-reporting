# To make "z-schema" available: yarn global add z-schema

lint: schemalint examplelint apilint

schemalint:
	z-schema configuration.json
	z-schema configuration-list.json
	z-schema tables-schema.json
	z-schema columns-schema.json
	z-schema query-schema.json
	z-schema results-schema.json
	z-schema template-query-schema.json
	z-schema template-results-schema.json

examplelint:
	z-schema configuration.json examples/configuration.json
	z-schema configuration-list.json examples/configuration-list.json
	z-schema tables-schema.json examples/tables-example.json
	z-schema columns-schema.json examples/columns-example.json
	z-schema query-schema.json examples/query-example.json
	z-schema results-schema.json examples/results-example.json
	z-schema template-query-schema.json examples/template-query-example.json
	z-schema template-results-schema.json examples/template-results-example.json

apilint: ldp.raml
	api_lint.py -t RAML -d .

doc:
	api_doc.py -o doc -t RAML -d .

