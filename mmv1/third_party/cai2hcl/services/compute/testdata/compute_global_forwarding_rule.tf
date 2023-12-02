resource "google_compute_global_forwarding_rule" "fr-1" {
  ip_protocol           = "TCP"
  ip_version            = "IPV4"
  load_balancing_scheme = "EXTERNAL_MANAGED"
  name                  = "fr-1"
  port_range            = "443"
  target                = "projects/myproj/global/targetSslProxies/tp-1"
}
