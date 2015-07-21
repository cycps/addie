FORMAT: 1A

# Cypress Design API
The API for designing [Cypress](https://cypress.deterlab.net) experiments.

## Experiment Update [/design/{xpid}]

+ Parameters
  + xpid (required, `system47`)

### Experiment Update [POST]

+ Request Update an experiment with a new computer (application/json)

        { 
          "Computers": [
            {
              "Name": "abby",
              "Sys": "system47",
              "Os": "Ubuntu1504-64-STD",
              "Start_script": "cook_muffins.sh"
            }
          ]
        }

+ Response 200 (application/json)

        {
          "Result": "ok",
          "Details": "",
          "Created": [
            {"Name": "abby", "Sys": "system47"}
          ]
        }

+ Request Update the computer abby with a different OS (application/json)

        { 
          "Computers": [
            {
              "Name": "abby",
              "Sys": "system47",
              "Os": "Debian-Sid",
              "Start_script": "cook_muffins.sh"
            }
          ]
        }

+ Response 200 (application/json)

        {
          "Result": "ok",
          "Updated": [
            {"Name": "abby", "Sys": "system47"}
          ]
        }

+ Request Update the computer abby in a non-existant system (application/json)

        { 
          "Computers": [
            {
              "Name": "abby",
              "Sys": "fake.system",
              "Os": "Debian-Sid"
            }
          ]
        }

+ Response 200 (application/json)

        {
          "Result": "failed",
          "Failed": [
            {
              "Name": "abby", 
              "Sys": "fake.system", 
              "Msg": "Non-existant system"
            }
          ]
        }


## Experiment Delete [/design/{xpid}/delete]
+ Parameters
  + xpid (required, `system47`)

### Experiment Delete[POST]
+ Request Update an experiment with a new computer (application/json)

        { 
          "Elements": [
            {
              "Name": "abby",
              "Sys": "system47"
            }
          ]
        }

+ Response 200 (application/json)

        {
          "Result": "ok",
          "Details": "",
          "Deleted": [
            {"Name": "abby", "Sys": "system47"}
          ]
        }
