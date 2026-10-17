ALTER TABLE ssh_servers
    ADD COLUMN host_key_fingerprint TEXT NOT NULL DEFAULT '';
