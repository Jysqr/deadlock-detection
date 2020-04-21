package DeadlockSite

import (
	"github.com/perlin-network/noise"
)

type Site struct {
	siteNodeCount  int
	TotalNodeCount int
	NodeList       []*DeadlockNode
}

func NewSite(bossNodeAddr string, numNode int, totalNode int) *Site {
	site := &Site{
		siteNodeCount:  numNode,
		TotalNodeCount: totalNode,
		NodeList:       make([]*DeadlockNode, numNode),
	}
	for i := 0; i < numNode; i++ { //builds the nodes and gives them their init parameters
		var dependence noise.ID
		if i > 0 {
			dependence = site.NodeList[i-1].node.ID()
		} else {
			dependence = noise.ID{
				ID:      noise.PublicKey{},
				Host:    nil,
				Port:    0,
				Address: "",
			}
		}
		dn := NewDeadlockNode(bossNodeAddr, numNode, site, dependence)
		go func(node *DeadlockNode) {
			dn.Start()
		}(dn)
		site.NodeList = append(site.NodeList, dn)
	}
	for i, siteNode := range site.NodeList {
		if i < len(site.NodeList) {
			siteNode.dependsOnMe = append(siteNode.dependsOnMe, site.NodeList[i+1].node.ID())
		}
	}

	return site
}
