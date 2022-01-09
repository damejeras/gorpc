package format

import (
	"bytes"
	"encoding/json"
	"go/doc"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/fatih/structtag"
	"github.com/pkg/errors"
)

func LoadTemplateFile(path string) (*template.Template, error) {
	return template.New(filepath.Base(path)).Funcs(map[string]interface{}{
		"camelize_down":       CamelizeDown,
		"camelize_up":         CamelizeUp,
		"json":                toJSONHelper,
		"format_comment_line": commentLine,
		"format_comment_text": commentText,
		"format_comment_html": commentHTML,
		"format_tags":         tags,
	}).ParseFiles(path)
}

func toJSONHelper(v interface{}) (template.HTML, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return "", err
	}
	return template.HTML(b), nil
}

func commentLine(s string) template.HTML {
	var buf bytes.Buffer
	doc.ToText(&buf, s, "", "", 2000)
	s = strings.TrimSpace(buf.String())
	return template.HTML(s)
}

func commentText(s string) template.HTML {
	var buf bytes.Buffer
	doc.ToText(&buf, s, "// ", "", 80)
	return template.HTML(buf.String())
}

func commentHTML(s string) template.HTML {
	var buf bytes.Buffer
	doc.ToHTML(&buf, s, nil)
	return template.HTML(buf.String())
}

// tags formats a list of struct tag strings into one.
// Will return an error if any of the tag strings are invalid.
func tags(tags ...string) (template.HTML, error) {
	alltags := &structtag.Tags{}
	for _, tag := range tags {
		theseTags, err := structtag.Parse(tag)
		if err != nil {
			return "", errors.Wrapf(err, "parse tags: `%s`", tag)
		}
		for _, t := range theseTags.Tags() {
			alltags.Set(t)
		}
	}
	tagsStr := alltags.String()
	if tagsStr == "" {
		return "", nil
	}
	tagsStr = "`" + tagsStr + "`"
	return template.HTML(tagsStr), nil
}

func isAcronym(word string) bool {
	for _, ac := range baseAcronyms {
		if strings.EqualFold(ac, word) {
			return true
		}
	}
	return false
}

var baseAcronyms = strings.Split(`HTML,JSON,JWT,ID,UUID,SQL,ACK,ACL,ADSL,AES,ANSI,API,ARP,ATM,BGP,BSS,CAT,CCITT,CHAP,CIDR,CIR,CLI,CPE,CPU,CRC,CRT,CSMA,CMOS,DCE,DEC,DES,DHCP,DNS,DRAM,DSL,DSLAM,DTE,DMI,EHA,EIA,EIGRP,EOF,ESS,FCC,FCS,FDDI,FTP,GBIC,gbps,GEPOF,HDLC,HTTP,HTTPS,IANA,ICMP,IDF,IDS,IEEE,IETF,IMAP,IP,IPS,ISDN,ISP,kbps,LACP,LAN,LAPB,LAPF,LLC,MAC,MAN,Mbps,MC,MDF,MIB,MoCA,MPLS,MTU,NAC,NAT,NBMA,NIC,NRZ,NRZI,NVRAM,OSI,OSPF,OUI,PAP,PAT,PC,PIM,PIM,PCM,PDU,POP3,POP,POTS,PPP,PPTP,PTT,PVST,RADIUS,RAM,RARP,RFC,RIP,RLL,ROM,RSTP,RTP,RCP,SDLC,SFD,SFP,SLARP,SLIP,SMTP,SNA,SNAP,SNMP,SOF,SRAM,SSH,SSID,STP,SYN,TDM,TFTP,TIA,TOFU,UDP,URL,URI,USB,UTP,VC,VLAN,VLSM,VPN,W3C,WAN,WEP,WiFi,WPA,WWW`, ",")
