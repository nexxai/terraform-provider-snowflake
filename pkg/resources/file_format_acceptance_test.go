package resources_test

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFileFormat_defaults(t *testing.T) {
	types := map[string]map[string]string{
		"csv": {
			"compression":                    "AUTO",
			"trim_space":                     "false",
			"record_delimiter":               "\n",
			"field_delimiter":                ",",
			"file_extension":                 "",
			"skip_header":                    "0",
			"skip_blank_lines":               "false",
			"date_format":                    "AUTO",
			"time_format":                    "AUTO",
			"timestamp_format":               "AUTO",
			"binary_format":                  "HEX",
			"escape":                         "",
			"escape_unenclosed_field":        "\\",
			"field_optionally_enclosed_by":   "",
			"error_on_column_count_mismatch": "true",
			"replace_invalid_characters":     "false",
			"validate_utf8":                  "true",
			"empty_field_as_null":            "true",
			"skip_byte_order_mark":           "true",
			"encoding":                       "UTF8",
			"null_if.#":                      "1",
			"null_if.0":                      "\\N",
		},
		"json": {
			"compression":      "AUTO",
			"trim_space":       "false",
			"date_format":      "AUTO",
			"time_format":      "AUTO",
			"timestamp_format": "AUTO",
			"binary_format":    "HEX",
			// docs say default should be same as csv above, but that's not observed
			"null_if.#":                  "0",
			"replace_invalid_characters": "false",
			"skip_byte_order_mark":       "true",
			"enable_octal":               "false",
			"allow_duplicate":            "false",
			"strip_outer_array":          "false",
			"strip_null_values":          "false",
			"ignore_utf8_errors":         "false",
		},
		"avro": {
			"compression": "AUTO",
			"trim_space":  "false",
			// docs say default should be same as csv above, but that's not observed
			"null_if.#": "0",
		},
		"orc": {
			"trim_space": "false",
			// docs say default should be same as csv above, but that's not observed
			"null_if.#": "0",
		},
		"parquet": {
			"compression":    "AUTO",
			"trim_space":     "false",
			"binary_as_text": "true",
			// docs say default should be same as csv above, but that's not observed
			"null_if.#": "0",
		},
		"xml": {
			"compression":        "AUTO",
			"trim_space":         "false",
			"ignore_utf8_errors": "false",
			// docs say default should be same as csv above, but that's not observed
			"null_if.#":              "0",
			"skip_byte_order_mark":   "true",
			"preserve_space":         "false",
			"strip_outer_element":    "false",
			"disable_snowflake_data": "false",
			"disable_auto_convert":   "false",
		},
	}

	for ttype, params := range types {
		t.Run(ttype, func(t *testing.T) {
			db := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
			schema := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
			name := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

			checks := []resource.TestCheckFunc{}

			for k, v := range params {
				checks = append(checks, resource.TestCheckResourceAttr("snowflake_file_format.ff", fmt.Sprintf("%s.0.%s", ttype, k), v))
			}
			resource.ParallelTest(t, resource.TestCase{
				Providers: providers(),
				Steps: []resource.TestStep{
					{
						Config: ffConfig(db, schema, name, ttype),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("snowflake_database.d", "name", db),
							resource.TestCheckResourceAttr("snowflake_schema.s", "name", schema),
							resource.TestCheckResourceAttr("snowflake_file_format.ff", "type", strings.ToUpper(ttype)),
							resource.ComposeTestCheckFunc(checks...),
						),
					},
					// IMPORT
					{
						ResourceName:      "snowflake_file_format.ff",
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

func TestAccFileFormat_changeType(t *testing.T) {
	name := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		Providers: providers(),
		Steps: []resource.TestStep{
			{
				Config: ffConfig(name, name, name, "csv"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database.d", "name", name),
					resource.TestCheckResourceAttr("snowflake_schema.s", "name", name),
					resource.TestCheckResourceAttr("snowflake_file_format.ff", "type", strings.ToUpper("csv")),
				),
			},
			{
				Config:      ffConfig(name, name, name, "json"),
				ExpectError: regexp.MustCompile("cannot change format type"),
			},
		},
	})
}

func TestAccFileFormat_rename(t *testing.T) {
	db := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	schema := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	name := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		Providers: providers(),
		Steps: []resource.TestStep{
			{
				Config: ffConfig(db, schema, name, "csv"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database.d", "name", db),
					resource.TestCheckResourceAttr("snowflake_schema.s", "name", schema),
					resource.TestCheckResourceAttr("snowflake_file_format.ff", "name", name),
					resource.TestCheckResourceAttr("snowflake_file_format.ff", "type", strings.ToUpper("csv")),
				),
			},
		},
	})
}

func ffConfig(db, schema, name, ttype string) string {
	s := `
resource snowflake_database d {
	name = "%s"
}

resource snowflake_schema s {
	database = snowflake_database.d.name
	name = "%s"
}

resource snowflake_file_format ff {
	database = snowflake_database.d.name
	schema = snowflake_schema.s.name
	name = "%s"
	
	%s {}
}
`
	s = fmt.Sprintf(s, db, schema, name, ttype)
	log.Println(s)
	return s
}