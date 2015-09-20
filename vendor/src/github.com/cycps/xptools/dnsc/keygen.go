package dnsc

import (
	"os/exec"
	"regexp"
)

func Keygen(dnsname string) (string, error) {
	cmd := exec.Command(
		"dnssec-keygen", "-r", "/dev/urandom",
		"-a", "HMAC-MD5", "-b", "512",
		"-n", "USER",
		dnsname+".")
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	kf, err := exec.Command("bash", "-c", "cat K"+dnsname+"*.private").Output()

	rx, _ := regexp.Compile("Key:\\s+(\\S+)")
	m := rx.FindStringSubmatch(string(kf))

	exec.Command("bash", "-c", "rm K"+dnsname+"*").Run()

	return m[len(m)-1], nil
}
