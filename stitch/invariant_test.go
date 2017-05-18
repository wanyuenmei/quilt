package stitch

import (
	"testing"
)

func TestReach(t *testing.T) {
	stc := `{
        "Containers": [
                {
                        "ID": "54be1283e837c6e40ac79709aca8cdb8ec5f31f5",
                        "Image": {"Name": "ubuntu"}
                },
                {
                        "ID": "3c1a5738512a43c3122608ab32dbf9f84a14e5f9",
                        "Image": {"Name": "ubuntu"}
                },
                {
                        "ID": "cb129f8a27df770b1dac70955c227a57bc5c4af6",
                        "Image": {"Name": "ubuntu"}
                }
        ],
        "Labels": [
                {
                        "Name": "a",
                        "IDs": ["54be1283e837c6e40ac79709aca8cdb8ec5f31f5"]
                },
                {
                        "Name": "b",
                        "IDs": ["3c1a5738512a43c3122608ab32dbf9f84a14e5f9"]
                },
                {
                        "Name": "c",
                        "IDs": ["cb129f8a27df770b1dac70955c227a57bc5c4af6"]
                }
        ],
        "Connections": [
                {"From": "a", "To": "b", "MinPort": 22, "MaxPort": 22},
                {"From": "b", "To": "c", "MinPort": 22, "MaxPort": 22}
        ],
        "Invariants": [
                {
                        "Form": "reach",
                        "Target": true,
                        "Nodes": ["a", "c"]
                },
                {
                        "Form": "reach",
                        "Target": false,
                        "Nodes": ["c", "a"]
                },
                {
                        "Form": "between",
                        "Target": true,
                        "Nodes": ["a", "c", "b"]
                },
                {
                        "Form": "between",
                        "Target": false,
                        "Nodes": ["c", "a", "b"]
                }
        ]
}`
	_, err := FromJSON(stc)
	if err != nil {
		t.Error(err)
	}
}

func TestReachPublic(t *testing.T) {
	stc := `{
        "Containers": [
                {
                        "ID": "54be1283e837c6e40ac79709aca8cdb8ec5f31f5",
                        "Image": {"Name": "ubuntu"}
                },
                {
                        "ID": "3c1a5738512a43c3122608ab32dbf9f84a14e5f9",
                        "Image": {"Name": "ubuntu"}
                },
                {
                        "ID": "cb129f8a27df770b1dac70955c227a57bc5c4af6",
                        "Image": {"Name": "ubuntu"}
                }
        ],
        "Labels": [
                {
                        "Name": "a",
                        "IDs": ["54be1283e837c6e40ac79709aca8cdb8ec5f31f5"]
                },
                {
                        "Name": "b",
                        "IDs": ["3c1a5738512a43c3122608ab32dbf9f84a14e5f9"]
                },
                {
                        "Name": "c",
                        "IDs": ["cb129f8a27df770b1dac70955c227a57bc5c4af6"]
                }
        ],
        "Connections": [
                {"From": "a", "To": "public", "MinPort": 22, "MaxPort": 22},
                {"From": "b", "To": "c", "MinPort": 22, "MaxPort": 22},
                {"From": "public", "To": "b", "MinPort": 22, "MaxPort": 22}
        ],
        "Invariants": [
                {
                        "Form": "reach",
                        "Target": true,
                        "Nodes": ["public", "b"]
                },
                {
                        "Form": "reach",
                        "Target": true,
                        "Nodes": ["public", "c"]
                },
                {
                        "Form": "reach",
                        "Target": false,
                        "Nodes": ["public", "a"]
                },
                {
                        "Form": "reach",
                        "Target": false,
                        "Nodes": ["b", "public"]
                }
        ]
	}`
	_, err := FromJSON(stc)
	if err != nil {
		t.Error(err)
	}
}

func TestNeighbor(t *testing.T) {
	stc := `{
        "Containers": [
                {
                        "ID": "54be1283e837c6e40ac79709aca8cdb8ec5f31f5",
                        "Image": {"Name": "ubuntu"}
                },
                {
                        "ID": "3c1a5738512a43c3122608ab32dbf9f84a14e5f9",
                        "Image": {"Name": "ubuntu"}
                },
                {
                        "ID": "cb129f8a27df770b1dac70955c227a57bc5c4af6",
                        "Image": {"Name": "ubuntu"}
                }
        ],
        "Labels": [
                {
                        "Name": "a",
                        "IDs": ["54be1283e837c6e40ac79709aca8cdb8ec5f31f5"]
                },
                {
                        "Name": "b",
                        "IDs": ["3c1a5738512a43c3122608ab32dbf9f84a14e5f9"]
                },
                {
                        "Name": "c",
                        "IDs": ["cb129f8a27df770b1dac70955c227a57bc5c4af6"]
                }
        ],
        "Connections": [
                {"From": "a", "To": "b", "MinPort": 22, "MaxPort": 22},
                {"From": "b", "To": "c", "MinPort": 22, "MaxPort": 22}
        ],
        "Invariants": [
                {
                        "Form": "reachDirect",
                        "Target": false,
                        "Nodes": ["a", "c"]
                },
                {
                        "Form": "reachDirect",
                        "Target": true,
                        "Nodes": ["b", "c"]
                }
        ]
	}`
	_, err := FromJSON(stc)
	if err != nil {
		t.Error(err)
	}
}

func TestAnnotation(t *testing.T) {
	stc := `{
        "Containers": [
                {
                        "ID": "54be1283e837c6e40ac79709aca8cdb8ec5f31f5",
                        "Image": {"Name": "ubuntu"}
                },
                {
                        "ID": "3c1a5738512a43c3122608ab32dbf9f84a14e5f9",
                        "Image": {"Name": "ubuntu"}
                },
                {
                        "ID": "cb129f8a27df770b1dac70955c227a57bc5c4af6",
                        "Image": {"Name": "ubuntu"}
                }
        ],
        "Labels": [
                {
                        "Name": "a",
                        "IDs": ["54be1283e837c6e40ac79709aca8cdb8ec5f31f5"]
                },
                {
                        "Name": "b",
                        "IDs": ["3c1a5738512a43c3122608ab32dbf9f84a14e5f9"],
                        "Annotations": ["ACL"]
                },
                {
                        "Name": "c",
                        "IDs": ["cb129f8a27df770b1dac70955c227a57bc5c4af6"]
                }
        ],
        "Connections": [
                {"From": "a", "To": "b", "MinPort": 22, "MaxPort": 22},
                {"From": "b", "To": "c", "MinPort": 22, "MaxPort": 22}
        ],
        "Invariants": [
                {
                        "Form": "reachACL",
                        "Target": false,
                        "Nodes": ["a", "c"]
                }
        ]
	}`

	_, err := FromJSON(stc)
	if err != nil {
		t.Error(err)
	}
}

func TestFail(t *testing.T) {
	stc := `{
        "Containers": [
                {
                        "ID": "54be1283e837c6e40ac79709aca8cdb8ec5f31f5",
                        "Image": {"Name": "ubuntu"}
                },
                {
                        "ID": "3c1a5738512a43c3122608ab32dbf9f84a14e5f9",
                        "Image": {"Name": "ubuntu"}
                },
                {
                        "ID": "cb129f8a27df770b1dac70955c227a57bc5c4af6",
                        "Image": {"Name": "ubuntu"}
                }
        ],
        "Labels": [
                {
                        "Name": "a",
                        "IDs": ["54be1283e837c6e40ac79709aca8cdb8ec5f31f5"]
                },
                {
                        "Name": "b",
                        "IDs": ["3c1a5738512a43c3122608ab32dbf9f84a14e5f9"]
                },
                {
                        "Name": "c",
                        "IDs": ["cb129f8a27df770b1dac70955c227a57bc5c4af6"]
                }
        ],
        "Connections": [
                {"From": "a", "To": "b", "MinPort": 22, "MaxPort": 22},
                {"From": "b", "To": "c", "MinPort": 22, "MaxPort": 22}
        ],
        "Invariants": [
                {
                        "Form": "reach",
                        "Target": true,
                        "Nodes": ["a", "c"]
                },
                {
                        "Form": "reach",
                        "Target": true,
                        "Nodes": ["c", "a"]
                }
        ]
	}`
	expectedFailure := `invariant failed: reach true "c" "a"`
	if _, err := FromJSON(stc); err == nil {
		t.Errorf("got no error, expected %s", expectedFailure)
	} else if err.Error() != expectedFailure {
		t.Errorf("got error %s, expected %s", err, expectedFailure)
	}
}

func TestBetween(t *testing.T) {
	stc := `{
        "Containers": [
                {
                        "ID": "54be1283e837c6e40ac79709aca8cdb8ec5f31f5",
                        "Image": {"Name": "ubuntu"}
                },
                {
                        "ID": "3c1a5738512a43c3122608ab32dbf9f84a14e5f9",
                        "Image": {"Name": "ubuntu"}
                },
                {
                        "ID": "cb129f8a27df770b1dac70955c227a57bc5c4af6",
                        "Image": {"Name": "ubuntu"}
                },
                {
                        "ID": "a0adbaafe75c74d1394e288855892f07f7463a0d",
                        "Image": {"Name": "ubuntu"}
                },
                {
                        "ID": "026453d646f6ffa2b834fcd84d58bfea2efac512",
                        "Image": {"Name": "ubuntu"}
                }
        ],
        "Labels": [
                {
                        "Name": "a",
                        "IDs": ["54be1283e837c6e40ac79709aca8cdb8ec5f31f5"]
                },
                {
                        "Name": "b",
                        "IDs": ["3c1a5738512a43c3122608ab32dbf9f84a14e5f9"]
                },
                {
                        "Name": "c",
                        "IDs": ["cb129f8a27df770b1dac70955c227a57bc5c4af6"]
                },
                {
                        "Name": "d",
                        "IDs": ["a0adbaafe75c74d1394e288855892f07f7463a0d"]
                },
                {
                        "Name": "e",
                        "IDs": ["026453d646f6ffa2b834fcd84d58bfea2efac512"]
                }
        ],
        "Connections": [
                {"From": "a", "To": "b", "MinPort": 22, "MaxPort": 22},
                {"From": "a", "To": "c", "MinPort": 22, "MaxPort": 22},
                {"From": "b", "To": "d", "MinPort": 22, "MaxPort": 22},
                {"From": "c", "To": "d", "MinPort": 22, "MaxPort": 22},
                {"From": "d", "To": "e", "MinPort": 22, "MaxPort": 22}
        ],
        "Invariants": [
                {
                        "Form": "reach",
                        "Target": true,
                        "Nodes": ["a", "e"]
                },
                {
                        "Form": "between",
                        "Target": true,
                        "Nodes": ["a", "e", "d"]
                }
        ]
}`
	_, err := FromJSON(stc)
	if err != nil {
		t.Error(err)
	}
}

func TestNoConnect(t *testing.T) {
	t.Skip("wait for scheduler, use the new scheduling algorithm")
	stc := `(label "a" (docker "ubuntu"))
(label "b" (docker "ubuntu"))
(label "c" (docker "ubuntu"))
(label "d" (docker "ubuntu"))
(label "e" (docker "ubuntu"))

(let ((cfg (list (provider "Amazon")
                 (region "us-west-1")
                 (size "m4.2xlarge")
                 (diskSize 32))))
    (makeList 4 (machine (role "test") cfg)))

(place (labelRule "exclusive" "e") "b" "d")
(place (labelRule "exclusive" "c") "b" "d" "e")
(place (labelRule "exclusive" "a") "c" "d" "e")

(invariant enough)`
	_, err := FromJSON(stc)
	if err != nil {
		t.Error(err)
	}
}

func TestNested(t *testing.T) {
	t.Skip("needs hierarchical labeling to pass")
	stc := `(label "a" (docker "ubuntu"))
(label "b" (docker "ubuntu"))
(label "c" (docker "ubuntu"))
(label "d" (docker "ubuntu"))

(label "g1" "a" "b")
(label "g2" "c" "d")

(connect 22 "g1" "g2")

(invariant reach true "a" "d")
(invariant reach true "b" "c")`
	_, err := FromJSON(stc)
	if err != nil {
		t.Error(err)
	}
}

func TestPlacementInvs(t *testing.T) {
	t.Skip("wait for scheduler, use the new scheduling algorithm")
	stc := `(label "a" (docker "ubuntu"))
(label "b" (docker "ubuntu"))
(label "c" (docker "ubuntu"))
(label "d" (docker "ubuntu"))
(label "e" (docker "ubuntu"))

(connect 22 "a" "b")
(connect 22 "a" "c")
(connect 22 "b" "d")
(connect 22 "c" "d")
(connect 22 "d" "e")
(connect 22 "c" "e")

(let ((cfg (list (provider "Amazon")
                 (region "us-west-1")
                 (size "m4.2xlarge")
                 (diskSize 32))))
    (makeList 4 (machine (role "test") cfg)))

(place (labelRule "exclusive" "e") "b" "d")
(place (labelRule "exclusive" "c") "b" "d" "e")
(place (labelRule "exclusive" "a") "c" "d" "e")

(invariant reach true "a" "e")
(invariant enough)`
	_, err := FromJSON(stc)
	if err != nil {
		t.Error(err)
	}
}
