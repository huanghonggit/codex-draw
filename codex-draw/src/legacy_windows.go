package main

func openLegacySecret(s storedSecret) (string, error) { return (platformVault{}).Open(s) }
