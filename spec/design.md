FORMAT: 1A

# Cypress Design API
The API for designing [Cypress](https://cypress.deterlab.net) experiments.

## Experiment Update [/design/{xpid}]
Use this endpoint to perform vairous design tasks on the experiment with the id *xpid*. See the docs for the request types associated with this endpoing for their semantics
+ Parameters
  + xpid (required, `test`)

### Experiment Update [POST]
Use this endpoint to update an experiment with the id *xpid*. The body of the request must contain an experiment element. If the provided element already exists it is updated. If the provided element does not exist it is created.

The response returns a list of the elements that have been updated in a list whose name represents the mode of update. In the examples below, the first response conatins an array called 'created' because the computer abby was created. In the second example the response contains an array called 'updated' because the computer abby was updated. The possible values of *result* are [ok,failed]. There is also an array called 'failed' that is possible if there were failures. If any element failed the whole operation fails and is rolled back. The response can also include lists for updated and failed elements. The following examples demonstrate this.


+ Request Update an experiment with a new computer (application/json)

        { 
          "computers": [
            {
              "name": "abby",
              "sys": "",
              "os": "Ubuntu1504-64-STD",
              "start_script": "cook_muffins.sh"
            }
          ]
        }

+ Response 200 (application/json)

        {
          "result": "ok",
          "created": [
            {"name": "abby", "sys": ""}
          ]
        }

<!--

+ Request Update the computer abby with a different OS (application/json)

        { 
          "computers": [
            {
              "name": "abby",
              "sys": "",
              "os": "Debian-Sid"
            }
          ]
        }

+ Response 200 (application/json)

        {
          "result": "ok",
          "updated": [
            {"name": "abby", "sys": ""}
          ]
        }

+ Request Update the computer abby in a non-existant system (application/json)

        { 
          "computers": [
            {
              "name": "abby",
              "sys": "fake.system",
              "os": "Debian-Sid"
            }
          ]
        }

+ Response 200 (application/json)

        {
          "result": "failed",
          "updated": [
            {
              "name": "abby", 
              "sys": "fake.system", 
              "msg": "The system 'fake.system' does not exist"
            }
          ]
        }
-->
