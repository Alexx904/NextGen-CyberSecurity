-- Host identity pins are lost; re-enroll them before connecting after re-upgrade.
ALTER TABLE ssh_servers DROP COLUMN host_key_fingerprint;
