(define Namespace "Mei")
(define Provider "Amazon")
(define AdminACL (list "local"))

(define MasterCount 1)
(define WorkerCount (+ 1 MasterCount))

(define machineCfg (list (provider Provider) (githubKey "yuenmeiwan")))

(makeList MasterCount (machine (role "Master") machineCfg))
(makeList WorkerCount (machine (role "Worker") machineCfg))

(label "containers" (makeList 3 (docker "ubuntu")))

(label "database" (docker "postgres"))

// Create 5 Apache containers, and label them "webTier"
(label "webTier" (makeList 5 (docker "httpd")))

// Create 2 Spark containers, and label them "batch"
(label "batch" (makeList 2 (docker "spark")))

// A deployment consists of a database, a webTier, and a batch processing
(label "deployment" (list "database" "webTier" "batch"))

// Allow the public internet to connect to the webTier over port 80
(connect 80 "public" "webTier")

// Allow the webTier to connect to the database on port 1433
(connect 1433 "webTier" "database")

// Allow the batch processer to connect to the database on and the webTier via SSH
(connect 22 "batch" (list "webTier" "database"))

// Allow all containers in the webTier to connect to each other on any port
(connect (list 0 65535) "webTier" "webTier")
