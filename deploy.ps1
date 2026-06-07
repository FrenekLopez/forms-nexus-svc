Write-Host "Starting isolated build with Docker..." -ForegroundColor Cyan

# 1. Build the image (Docker reads the Dockerfile and compiles the code inside the container)
docker build -t forms-nexus-builder .

# Stop the process if Docker fails
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ ERROR: Build failed. Deployment has been aborted to prevent publishing an outdated version." -ForegroundColor Red
    exit 1
}

Write-Host "Extracting the Linux binary to the local environment..." -ForegroundColor Cyan

# 2. Ensure the local output directory exists
New-Item -ItemType Directory -Force -Path .\bin | Out-Null

# 3. Create a temporary container to extract the generated artifact
docker create --name temp-extractor forms-nexus-builder
docker cp temp-extractor:/app/bootstrap ./bin/bootstrap

# 4. Remove the temporary container to avoid leaving unused resources behind
docker rm temp-extractor

Write-Host "Binary ready. Starting deployment to AWS..." -ForegroundColor Green

# 5. Navigate to the infrastructure directory and run the deployment
cd deployments
cdk deploy
cd ..

Write-Host "Deployment completed successfully!" -ForegroundColor Yellow