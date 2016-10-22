package client

import "golang.org/x/net/context"

// Ping pings the server and return the value of the "Docker-Experimental" header
func (cli *Client) Ping(ctx context.Context) (bool, string, error) {
	serverResp, err := cli.get(context.WithValue(ctx, "ignore-version", true), "/_ping", nil, nil)
	if err != nil {
		return false, "", err
	}
	defer ensureReaderClosed(serverResp)

	apiVersion := serverResp.header.Get("API-Version")

	exp := serverResp.header.Get("Docker-Experimental")
	if exp != "true" {
		return false, apiVersion, nil
	}

	return true, apiVersion, nil
}
