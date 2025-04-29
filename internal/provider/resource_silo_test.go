// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// This self-signed TLS certificate and key are just for testing. It's not
// critical to anything nor is it a security risk.
// TODO: Configure the TLS certificate and key in another way to prevent static
// analysis tools from flagging this as a false positive.
const (
	tlsCertificateBase64 = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUdDVENDQS9HZ0F3SUJBZ0lVYjB2YzdFUkZGQ3ZtcGpuak5YREFVTmVsclBZd0RRWUpLb1pJaHZjTkFRRUwKQlFBd2daTXhDekFKQmdOVkJBWVRBbFZUTVJNd0VRWURWUVFJREFwRFlXeHBabTl5Ym1saE1STXdFUVlEVlFRSApEQXBGYldWeWVYWnBiR3hsTVI4d0hRWURWUVFLREJaUGVHbGtaU0JEYjIxd2RYUmxjaUJEYjIxd1lXNTVNUlF3CkVnWURWUVFMREF0RmJtZHBibVZsY21sdVp6RWpNQ0VHQTFVRUF3d2FLaTV6ZVhNdWNqTXViM2hwWkdVdGNISmwKZG1sbGR5NWpiMjB3SGhjTk1qVXdOREkyTURJME1UUTVXaGNOTWpZd05ESTJNREkwTVRRNVdqQ0JrekVMTUFrRwpBMVVFQmhNQ1ZWTXhFekFSQmdOVkJBZ01Da05oYkdsbWIzSnVhV0V4RXpBUkJnTlZCQWNNQ2tWdFpYSjVkbWxzCmJHVXhIekFkQmdOVkJBb01Gazk0YVdSbElFTnZiWEIxZEdWeUlFTnZiWEJoYm5reEZEQVNCZ05WQkFzTUMwVnUKWjJsdVpXVnlhVzVuTVNNd0lRWURWUVFEREJvcUxuTjVjeTV5TXk1dmVHbGtaUzF3Y21WMmFXVjNMbU52YlRDQwpBaUl3RFFZSktvWklodmNOQVFFQkJRQURnZ0lQQURDQ0Fnb0NnZ0lCQUptRU1oQU1Pbm52Qkd0UUhsbEJKWk1KCmgzTjR6Z3E2YjQ4cGJqOHErbGZtVzBmd092ek5MV3dvTDRtNGNkZkh0WE95RkpOU3dONHhhblNsM2krM09qREMKM1hZMURMZllNQi9sdzk1WlZVYktGdlZ4Z1pSWmpZVHN6eTR1emMrZ3hHejFrSGpMYlc2U2hlcG1DK01IblNEWApXZGhEUVBFSTY4WVVWN1lOWXJ5YlBxNXlFRWFvNlJCRkEvOTBqNHE1MHpSMVpRSExIL1BDcmYyTkt3VmhhUjZHCjE3U3NnZHg5YkJmcUFiM2tFMmxCUVZMdUpqMjRMWUc1ZUNSajd3SzRwREx2VXNJNkdSblpUanMwN24zSTVrWHkKSjVubWFERTB5amN0U1lFaWc4dTNIODcveGlxVEdtOVRIaHMrSHI5cHM1RGxmZFpINHNSUWNiYWZHb2lGUmQ1ZgpxandIMk9HdWJ6NXRGQ3JFV3JBMkRMaWtHYTVrMGxmL2svY2w0Zy9MdlpOZDJqcVdJQmxBdzM4RDl5OWFjbStBCkJjYjRoRlRXYjFwKzhuSDZBdyt1cGxQTkU2U2dLQTA3Ny95TG5UVnByQm9uaEs2OVQ1dWlvM1FOZG9hVUg5NkEKUkZQdklHWXJRTHpYN204eFpub0Fab2lSY21zVmhvZTVSRExhZFJIaUsrUFROSmhzdlRuTFJ5RUxMUnppTlpwRApyUitqMk5PZXVILzNnblBHdm5qVnJCZjRpUGYzY0I0WlZMajZRZE5ueHdUWUhXWS95Y096cGY0ODJoRk9QbzRBCkg3WUVtTzRTcGdvRndTdlFjbGZ3R3g3eW9Sb1Bzc2hmRzhJYW5CN2ZTZDlQYmFNemUwMzhEQkxVMk5qaDVlb28KQWlHSFB3OC92TnpHd2hoaXZvSExBZ01CQUFHalV6QlJNQjBHQTFVZERnUVdCQlJQRlhUdm83OHdwUWtLaHhDMgo0OXpnL01DSTZEQWZCZ05WSFNNRUdEQVdnQlJQRlhUdm83OHdwUWtLaHhDMjQ5emcvTUNJNkRBUEJnTlZIUk1CCkFmOEVCVEFEQVFIL01BMEdDU3FHU0liM0RRRUJDd1VBQTRJQ0FRQ0VmRFpzZTRIRW1PME03dWRjR2hHeWZpTXoKN3A2Y0lXQS9XVThaSEFLT04zSGlwYUpwalp4OE1aVkVaMzVobTRROFEzYkxoRHpZVTJmV3RMMXhBN0JyaHZCMQpudGN0VXRVOGtCN1JKbzl3VjU5UFREaFhzMzRiV1JyZ1lzZkRFL3ExZjNCQ2xRanhBTTA0T3hJeXVKSzNkendUCnBvbXdNaFpKSmRYcXhoL1NnR0IwdzhZUkw5WEV6dVh1aFJoNmk2b0N2Y0FQQ0VDcjVuQllraWlVSmJiams4YVMKdDVQMVJSTFdEV2MrSGloYWx4cmlFRjRUbFFUbW96N2cxdzZJVDdYZjFkZ1A1aTlaN1Vad1RWQlRnZG1naW5JVgp4SnFTOVJNN3pVSTJqWHU5aTk0Zk5qOVA4VUJWRysxS3BqazQvaE5VOWp2TGNGRFNPRjZkSnZwRmFmKzg1di9WCjJ2WFVMb1FJV3IvcDF2dUdMTmxnb0M1WFptN2ZzdDlRcEdva1hlMEwySWp0Q1pncWRKRzVBOHhiaDJEa1orbEQKNXk5akVMOTRGRXNBTndFelhxQlVDQWpjZ2pwUVJBajM2Nk9tb1pSVDFCeUNYK0RXNFBXRnJvclorc0Z6VStmNQpCQUpFaTJVNmg2VGdaQ2d1Ty81Z0l6QWtDL0xXdThYb0lwNExGVnVWRmhvQzJSZm40YWhxdm42RHNSdTVqcWoyCnhIa1RLUVdmZlN3dGQ5akRQWm9Db2JEYkFXREdOWllTQ2tyMU5icmZYOGRBL3cwK2ZXb1NMVjhiOEhzMG5CSUgKNWpkdkJOQTdVR1UzK0tnamFrenJSRVkwWnRJUEJGdEZZVGMxb3Y0aHQrZVpxeEQxcEh5VFArMHltYkJYZXlaeQpKMlVJRnQwUThQSzhOUkdXRkE9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
	tlsPrivateKeyBase64  = "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUpRZ0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQ1N3d2dna29BZ0VBQW9JQ0FRQ1poRElRRERwNTd3UnIKVUI1WlFTV1RDWWR6ZU00S3VtK1BLVzQvS3ZwWDVsdEg4RHI4elMxc0tDK0p1SEhYeDdWenNoU1RVc0RlTVdwMApwZDR2dHpvd3d0MTJOUXkzMkRBZjVjUGVXVlZHeWhiMWNZR1VXWTJFN004dUxzM1BvTVJzOVpCNHkyMXVrb1hxClpndmpCNTBnMTFuWVEwRHhDT3ZHRkZlMkRXSzhtejZ1Y2hCR3FPa1FSUVAvZEkrS3VkTTBkV1VCeXgvendxMzkKalNzRllXa2VodGUwcklIY2ZXd1g2Z0c5NUJOcFFVRlM3aVk5dUMyQnVYZ2tZKzhDdUtReTcxTENPaGtaMlU0NwpOTzU5eU9aRjhpZVo1bWd4Tk1vM0xVbUJJb1BMdHgvTy84WXFreHB2VXg0YlBoNi9hYk9RNVgzV1IrTEVVSEcyCm54cUloVVhlWDZvOEI5amhybTgrYlJRcXhGcXdOZ3k0cEJtdVpOSlgvNVAzSmVJUHk3MlRYZG82bGlBWlFNTi8KQS9jdlduSnZnQVhHK0lSVTFtOWFmdkp4K2dNUHJxWlR6Uk9rb0NnTk8rLzhpNTAxYWF3YUo0U3V2VStib3FOMApEWGFHbEIvZWdFUlQ3eUJtSzBDODErNXZNV1o2QUdhSWtYSnJGWWFIdVVReTJuVVI0aXZqMHpTWWJMMDV5MGNoCkN5MGM0aldhUTYwZm85alRucmgvOTRKenhyNTQxYXdYK0lqMzkzQWVHVlM0K2tIVFo4Y0UyQjFtUDhuRHM2WCsKUE5vUlRqNk9BQisyQkpqdUVxWUtCY0VyMEhKWDhCc2U4cUVhRDdMSVh4dkNHcHdlMzBuZlQyMmpNM3ROL0F3UwoxTmpZNGVYcUtBSWhoejhQUDd6Y3hzSVlZcjZCeXdJREFRQUJBb0lDQUNITDlLbUx4NlBvZHZTWkl0VkxmbFlzCmx1RlpDeU5aZ0Ezb2RSajdBVG93d0kvSjEzS29TUU95cFNTUXNwOXFuQXZvZkpjaWRNdDEzWlhvbmsycTdPaW4KUGRJMFE2U0Z0N0tPQnQwQWxjR0w1Qm9NN3hZVjBRNGVoRTRLaDh6WisrUncrMmxjZjY4RUd1OUxuL3BQUnN4ZwpIS3Q3d3VSTnJucGhLQjR3UERpQmhQOHFwV0tvOVFaYjYxRmwrK1B5blFqRGY0VXhqc3MvWk1hWk9ZdHBzcGJCCjRPTXB4ejBmYjVpa0w5WDZURHV6M2dtLzNETmlSTUoyYm5pMGQzNEY0RUJHWjlYU3JJd0FSelRKcG1lU3Z2OVAKSEdESlZNN2diRlJSYUFsQjYvb0JTc05yazlqem9iSTRmanhKSk1QSEpYMFV5T3RQMENDZ0JTakxSakFnQncxaQpmQmh4bS9pQmhveU50UFVWeGVRa2VQSE81d2lSZVQydDErTDZha0tmMlVXLzIrZTBvYlNqSGJsWkFWaGVRN3FJClcvREwyRURBZnFTMlZ4clVtY09xVzBHd21IQTFFMDRhRG5lOC9aZ1NoVVJ1Z2xlcndsMGhzeFdhcFAwNlFrWFUKYTN4TEpQZ1RwYzBJY25ySE0wdHNIaUhNY3E2KzJ3MFF0eHpuYkZ0em9wSkx1c01jUXhNSmxKdkFGcFkzdWZNWgpIMFUxY3RuQXprQmVWUVNvRE9SY1dSVGd0aFlZbllYMnpyb1NpeGlJZ1RrRWlXa25McGxsUWFrMjZvTWtRZFZmClB4L2lycjYvNUdTMTJsYm41WldCRTBUMXhyZzkxNXlRZ2NwamRneXNQTW1TbDRGUUtvSnBvbENSTE5taGdOZmcKUWpWV00xL1FZMUlGWDBILzZEcmhBb0lCQVFES1BtckJiOVlURXNTYTZadnFMdk5wRnpnWFUybU9JSE5MY2p3Uwozb0VzQUhXL2FITzhpZWNMUzBEazBBZ3dDNFJndGg0ckNtbllUbU5CVERCNjkwN2t3R3RGMTBMZ1dzeHA3L0I1CkxLeDhINGtIZzdCYWd3VExGSDgxUDFWL1I2OUJuS2NhZnZoVkpzYUxlUVVtekJnRndwQWNxbVBKMk5abVJjZE4KVG9USG9pMHF0c0VHai83TjBaakNWQWtmZUJHWU1mSFY3YVVBWDhRRDgrSWcyazFCNEduaHRhSHA2c1MvbktBNQpSWHFqM05QUU41bEtCd0FKVzIwUytiMjJHZmlaUnh0ZHFGY3JRZXRHUUZ2cElWSzJPMDNmNjZuYmpvNjNwdXZvCjFkejlFckduckFIT1VhU3dvd2tXaW8xc3VjN0d0TU42NzJDSHFNa1ZqbFJHUlFmM0FvSUJBUURDVWlUc0RlRGEKd2NuZnhDcE9UUStWbEtlRVVGSXNMM1czeTJteXdKeHRJd0UxcUtXTjFWYjFWczVja1dBWHppdVFlQ29lbEhPSQpISGthcWJaM0ZoeEZWdVdNemhOTURhUHlFZWFKY0JjVnhGN1BIY1VrOGd3V3BsR2xrTlZCYktpaHgvU0pxUnVqCmR0cVVTc1hyK2hRejlZNDhFTGdpSUpsN08zL04vMlRjcnpsWmpQcXdxT0RGaWVER2tpSlhFam5EcEJiTDNBSU4KdEE1RkZKV0lWc2pVN0dCNzFyRnJFZmpTT0IzWFpIc05hV3RXNnlvakhmUmE5Y2dEKzBLQytjSUdVRm5rZjI3VQpCOWdwTXZEdjkyMXdPZGh2akJpYzVvSEN2VCsxa0ZDL1Y0S00xcVBpazB0bEJNajBPR0tmeUlIWWdBT00xTGlvCjFhWVk4VFJWbzZmTkFvSUJBRWUzcXBPOTNPUVdtN0Z6ZGQ2dGw1T0VzRmRWTlBFNWdLa1ljVVVmc2g2d2F4RGQKTVcyQ1dYUWYwM3RRYWhiZmZxbnM2dlhJVTVCbys3bUVFdzBIOWVvWWNmSHFTOFRUYmZtREpIdFQ1RFovMkUvWgoyd2U5dmsxbGoxYUtodjhEcEpwWHVzb2lqRjFseXJKYTBBRGFEd3E3Mis3T1hXU09pRGpzTmFpc1YxbVRvUUNzCm5mWjl5WldpNWRERGpCaWtzMWlOSFgwSE1LUFpVZUUwOHRORGxuSHQ2cDRua3Fzb25XeDFWanY0NzJ4OE9vQnoKdHVBUmEybm1DZC9Zdi9WN2NEU3Fpb0hEMkdWMmtyL3V3cWtCTUJ0L0hEWnprMkJRUlR2SzdZMDdpWW9VdnZyKwpmQVYxM2pqbEY2dnVwZ2dRTzhzcS9zYnhiQUd2VU45Y0FYYUp0REVDZ2dFQkFJalRVbEFzVFlGN0ptdzdNaGJFClNBN3BCek14WTByZGVDUWNSS2FxM1BvenhheEV2WjJxOUhuM3o0SjZrcER3aU5oRzVGRjM4Z21MRXZMbFFTZUYKR0E3eTZ0dEVWMjRieEs2MFVBSENQVjhFVUVYQ0RvaS9MaWZjb0d6V0dITGkwYkpvbXhVN1Q4eS82WlMxT2J1NAo4UFROR0lQT3VmaTl2NVI0QnJ2RDh2ODVHa2FsNy9ib1VxeUZNeEplMzNNejBCeWpzN0dEanFhYmU5akViNjM2CmZaci9mY2gxR2FQc21hbGIvaGNtRjBjUVRaWjhLOFZpV0NhY0hXUkFUVXJ3RmVCZ1A0dVc5ekN3L1ZHMUh0VzUKQVFRZWx2bWtTY2hndmttaSsvTWFWT0VGKzFTejVkMnFIVkphRmkxd2JuRlh1Nlg0TFlmQ1dPdjQwK1dJSVhPVApzcGtDZ2dFQU5rbi9Pb1VpQmVReVRFQ3VKRlBJSjBXa2VoL0lKdW5BOHIzdkxObEV6ZlB0QjA5KzFyNm9NeTdPClpZNmQxL3VQRkxGRkNSMUtBaDlQdTZUUHpkSUcybFNOSVhYNm9GdmcyWENwcGZucXI0RmFJWHMrYzVzWWtMamoKSitTL0svRTBzOVg4cEtjL3FBLytmSXRtTW1oRFRaV0dOZEhOU3YzQzZLcnBkSXptdmhjQ2FHUjIxNWlkaW5KaApCeWR2M0hPZk5DMG5BLzFtZ2hoUTNOQXY1bWZ1cFBKVUR0S1RqSi94UHdqRmppNFBEem56UXE5VlRVYlZLcElOCko1dTNrZ3FDSmFrcFJVWTRFM2pndmN3bWhHUVlhNVpvUDFOVnhZUTQwcXdTM2VBNzhyaldoaGdralA1T1RmdEcKY3Z4WGNOUlVsN1owTGt6UWlCLzh4OWxHd2VYdXJBPT0KLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQo="
)

type resourceSiloConfig struct {
	BlockName            string
	SiloName             string
	TLSCertificateBase64 string
	TLSPrivateKeyBase64  string
}

var resourceSiloConfigTpl = `
resource "oxide_silo" "{{.BlockName}}" {
  name          = "{{.SiloName}}"
  description   = "Managed by Terraform."
  discoverable  = true
  identity_mode = "local_only"

  quotas = {
    cpus    = 2
    memory  = 8589934592
    storage = 8589934592
  }

  mapped_fleet_roles = {
    admin  = ["admin", "collaborator"]
    viewer = ["viewer"]
  }

  tls_certificates = [
    {
      name        = "self-signed-wildcard"
      description = "Self-signed wildcard certificate for *.sys.r3.oxide-preview.com."
      cert        = base64decode("{{.TLSCertificateBase64}}")
      key         = base64decode("{{.TLSPrivateKeyBase64}}")
      service     = "external_api"
    },
  ]

  timeouts = {
    create = "1m"
    read   = "2m"
    update = "3m"
    delete = "4m"
  }
}
`

var resourceSiloUpdateConfigTpl = `
resource "oxide_silo" "{{.BlockName}}" {
  name          = "{{.SiloName}}"
  description   = "Managed by Terraform."
  discoverable  = true
  identity_mode = "local_only"

  quotas = {
    cpus    = 4           # 2 -> 4
    memory  = 17179869184 # 8 GiB -> 16 GiB
    storage = 17179869184 # 8 GiB -> 16 GiB
  }

  mapped_fleet_roles = {
    admin  = ["admin", "collaborator"]
    viewer = ["viewer"]
  }

  tls_certificates = [
    {
      name        = "self-signed-wildcard"
      description = "Self-signed wildcard certificate for *.sys.r3.oxide-preview.com."
      cert        = base64decode("{{.TLSCertificateBase64}}")
      key         = base64decode("{{.TLSPrivateKeyBase64}}")
      service     = "external_api"
    },
  ]
}
`

func TestAccSiloResourceSilo_full(t *testing.T) {
	siloName := newResourceName()
	blockName := newBlockName("silo")
	resourceName := fmt.Sprintf("oxide_silo.%s", blockName)
	config, err := parsedAccConfig(
		resourceSiloConfig{
			BlockName:            blockName,
			SiloName:             siloName,
			TLSCertificateBase64: tlsCertificateBase64,
			TLSPrivateKeyBase64:  tlsPrivateKeyBase64,
		},
		resourceSiloConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configUpdate, err := parsedAccConfig(
		resourceSiloConfig{
			BlockName:            blockName,
			SiloName:             siloName,
			TLSCertificateBase64: tlsCertificateBase64,
			TLSPrivateKeyBase64:  tlsPrivateKeyBase64,
		},
		resourceSiloUpdateConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSiloDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceSilo(resourceName, siloName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceSiloUpdate(resourceName, siloName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceSilo(resourceName string, siloName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", siloName),
		resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform."),
		resource.TestCheckResourceAttr(resourceName, "quotas.cpus", "2"),
		resource.TestCheckResourceAttr(resourceName, "quotas.memory", "8589934592"),
		resource.TestCheckResourceAttr(resourceName, "quotas.storage", "8589934592"),
		resource.TestCheckResourceAttr(resourceName, "discoverable", "true"),
		resource.TestCheckResourceAttr(resourceName, "identity_mode", "local_only"),
		resource.TestCheckResourceAttrSet(resourceName, "mapped_fleet_roles.admin.0"),
		resource.TestCheckResourceAttrSet(resourceName, "mapped_fleet_roles.viewer.0"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "4m"),
	}...)
}

func checkResourceSiloUpdate(resourceName string, siloName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", siloName),
		resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform."),
		resource.TestCheckResourceAttr(resourceName, "quotas.cpus", "4"),
		resource.TestCheckResourceAttr(resourceName, "quotas.memory", "17179869184"),
		resource.TestCheckResourceAttr(resourceName, "quotas.storage", "17179869184"),
		resource.TestCheckResourceAttr(resourceName, "discoverable", "true"),
		resource.TestCheckResourceAttr(resourceName, "identity_mode", "local_only"),
		resource.TestCheckResourceAttrSet(resourceName, "mapped_fleet_roles.admin.0"),
		resource.TestCheckResourceAttrSet(resourceName, "mapped_fleet_roles.viewer.0"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func testAccSiloDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_silo" {
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		params := oxide.SiloViewParams{
			Silo: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}

		res, err := client.SiloView(ctx, params)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("silo (%v) still exists", &res.Name)
	}

	return nil
}
