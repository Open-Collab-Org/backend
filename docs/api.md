# Errors

When a request to the API results in a `4XX` error, the body of the response
will contain a json object describing the error that occurred.

```
{
  "errorCode": string,
  "errorDetails": object
}
```

- `errorCode`: A code indicating the type of error that exists.
- `errorDetails`: An object containing extra information about the error.

## Validation error

A validation error (unsurprisingly) has the code `validation-error`. The `errorDetails` is
a map that correlates field names with failed constraints. Only fields that had invalid
values will be present in the error object.

```json
{
  "errorCode": "validation-error",
  "errorDetails": {
    "fieldName": "constraint"
  }
}
```

Example:

The following error is a validation error that was caused because the value of the field
`longDescription` didn't match the constraint `min=200`, which means it was shorter
than the minimum length, which is 200.
```json
{
    "errorCode": "validation-error",
    "errorData": {
        "LongDescription": "min=200"
    }
}
```
