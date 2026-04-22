#!/bin/bash
echo "Waiting for mongo nodes to start..."

until mongosh --host mongo:27017 --eval 'quit(0)' &>/dev/null; do
  echo "Waiting for mongo..."
  sleep 2
done

echo "Initializing replica set..."

mongosh --host mongo:27017 <<EOF
try {
  rs.status()
} catch (e) {
  rs.initiate({
    _id: "rs0",
    members: [
      { _id: 0, host: "mongo:27017" },
    ]
  })
}
EOF

echo "Replica set initialized."