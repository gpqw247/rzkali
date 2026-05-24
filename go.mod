# Delete broken go.mod
rm go.mod
rm go.sum 2>/dev/null

# Create minimal correct go.mod
cat > go.mod << 'EOF'
module autorzp

go 1.21
EOF
