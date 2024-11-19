---
page_title: "genesyscloud_telephony_providers_edges_site_outbound_route Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Outbound Routes for a Genesys Cloud Site
---
# genesyscloud_telephony_providers_edges_site_outbound_route (Resource)

Outbound Routes for a Genesys Cloud Site

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

- [GET /api/v2/telephony/providers/edges/sites](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#get-api-v2-telephony-providers-edges-sites)
- [GET /api/v2/telephony/providers/edges/sites/{siteId}/outboundroutes](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#get-api-v2-telephony-providers-edges-sites--siteId--outboundroutes)
- [POST /api/v2/telephony/providers/edges/sites/{siteId}/outboundroutes](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#post-api-v2-telephony-providers-edges-sites--siteId--outboundroutes)
- [DELETE /api/v2/telephony/providers/edges/sites/{siteId}/outboundroutes/{outboundRouteId}](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#delete-api-v2-telephony-providers-edges-sites--siteId--outboundroutes--outboundRouteId-)
- [PUT /api/v2/telephony/providers/edges/sites/{siteId}/outboundroutes/{outboundRouteId}](https://developer.genesys.cloud/api/rest/v2/telephonyprovidersedge/#put-api-v2-telephony-providers-edges-sites--siteId--outboundroutes--outboundRouteId-)

#### Compatibility Note

In versions 1.39.0 to 1.48.0 of the provider, this resource was constructed with a different structure. The current version introduces structural changes that are not backwards compatible with those earlier versions.

These changes are currently controlled by a feature flag and are not yet the default behavior. This allows for a phased implementation and thorough testing of this resource before full release.

If you're upgrading from an earlier version, please be aware of these structural changes and consult these examples on how to migrate your configuration.


## Example Usage

```terraform
// To enable this resource, set ENABLE_STANDALONE_OUTBOUND_ROUTES as an environment variable
resource "genesyscloud_telephony_providers_edges_site_outbound_routes" "site1-route1" {
  site_id                 = genesyscloud_telephony_providers_edges_site.site1.id
  name                    = "outboundRoute 1"
  description             = "outboundRoute description"
  classification_types    = ["International", "National"]
  external_trunk_base_ids = [genesyscloud_telephony_providers_edges_trunkbasesettings.trunk-base-settings1.id]
  distribution            = "RANDOM"
  enabled                 = false
}

resource "genesyscloud_telephony_providers_edges_site_outbound_routes" "site1-route2" {
  site_id                 = genesyscloud_telephony_providers_edges_site.site1.id
  name                    = "outboundRoute 2"
  description             = "outboundRoute description"
  classification_types    = ["Network"]
  external_trunk_base_ids = [genesyscloud_telephony_providers_edges_trunkbasesettings.trunk-base-settings2.id]
  distribution            = "SEQUENTIAL"
  enabled                 = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `classification_types` (List of String) Used to classify this outbound route.
- `name` (String) The name of the entity.
- `site_id` (String) The Id of the site to which the outbound routes belong.

### Optional

- `description` (String) The resource's description.
- `distribution` (String) Valid values: SEQUENTIAL, RANDOM. Defaults to `SEQUENTIAL`.
- `enabled` (Boolean) Enable or disable the outbound route Defaults to `false`.
- `external_trunk_base_ids` (List of String) Trunk base settings of trunkType "EXTERNAL". This base must also be set on an edge logical interface for correct routing. The order of the IDs determines the distribution if "distribution" is set to "SEQUENTIAL"

### Read-Only

- `id` (String) The ID of this resource.
- `route_id` (String) The Id of the outbound route. This is distinct from the "id" field. The "id" field is a combination of the site_id and route_id
