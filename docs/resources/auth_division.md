---
page_title: "genesyscloud_auth_division Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud Authorization Division
---
# genesyscloud_auth_division (Resource)

Genesys Cloud Authorization Division

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [GET /api/v2/authorization/divisions/home](https://developer.mypurecloud.com/api/rest/v2/authorization/#get-api-v2-authorization-divisions-home)
* [POST /api/v2/authorization/divisions](https://developer.mypurecloud.com/api/rest/v2/authorization/#post-api-v2-authorization-divisions)
* [GET /api/v2/authorization/divisions/{divisionId}](https://developer.mypurecloud.com/api/rest/v2/authorization/#get-api-v2-authorization-divisions--divisionId-)
* [GET /api/v2/authorization/divisions](https://developer.mypurecloud.com/api/rest/v2/authorization/#get-api-v2-authorization-divisions)
* [PUT /api/v2/authorization/divisions/{divisionId}](https://developer.mypurecloud.com/api/rest/v2/authorization/#put-api-v2-authorization-divisions--divisionId-)
* [DELETE /api/v2/authorization/divisions/{divisionId}](https://developer.mypurecloud.com/api/rest/v2/authorization/#delete-api-v2-authorization-divisions--divisionId-)


## Example Usage

```terraform
resource "genesyscloud_auth_division" "marketing" {
  name        = "Marketing"
  description = "Custom Division for Marketing"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **name** (String) Division name.

### Optional

- **description** (String) Division description.
- **home** (Boolean) True if this is the home division. This can be set to manage the pre-existing home division.
- **id** (String) The ID of this resource.
