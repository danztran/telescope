package collector

import "github.com/danztran/telescope/pkg/scope"

func getNodePod(node scope.APINode) string {
	for _, p := range node.Node.Parents {
		if p.TopologyID == "pods" {
			return p.Label
		}
	}
	return ""
}

func getPodUID(node scope.APINode) string {
	for _, lb := range node.Node.Tables {
		if lb.ID == scope.LabelDocker {
			for _, row := range lb.Rows {
				if row.ID == scope.LabelPodUID {
					return row.Entries["value"]
				}
			}
		}
	}
	return ""
}

func getOutgoingConnections(node scope.APINode) []scope.Connection {
	for _, c := range node.Node.Connections {
		if c.ID == scope.OutboundID {
			return c.Connections
		}
	}
	return nil
}

func getConnectionPort(conn scope.Connection) string {
	port := ""
	for _, meta := range conn.Metadata {
		if meta.ID == "port" {
			port = meta.Value
			break
		}
	}

	return port
}
