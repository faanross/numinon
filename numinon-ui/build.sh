#!/bin/bash

# Numinon UI Build Script
# This script prepares icons and builds the application

echo "🚀 Building Numinon UI..."

# Create build directory if it doesn't exist
mkdir -p build

# Check if icon exists in the tray directory
if [ -f "internal/tray/icon.png" ]; then
    echo "✅ Found icon.png in internal/tray/"

    # Copy icon to build directory
    cp internal/tray/icon.png build/appicon.png

    # For macOS, we need to create an icon set
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "🍎 Detected macOS - Creating icon set..."

        # Create iconset directory
        mkdir -p build/appicon.iconset

        # Generate different sizes (requires imagemagick: brew install imagemagick)
        if command -v sips &> /dev/null; then
            # Use macOS native sips tool
            sips -z 16 16     build/appicon.png --out build/appicon.iconset/icon_16x16.png
            sips -z 32 32     build/appicon.png --out build/appicon.iconset/icon_16x16@2x.png
            sips -z 32 32     build/appicon.png --out build/appicon.iconset/icon_32x32.png
            sips -z 64 64     build/appicon.png --out build/appicon.iconset/icon_32x32@2x.png
            sips -z 128 128   build/appicon.png --out build/appicon.iconset/icon_128x128.png
            sips -z 256 256   build/appicon.png --out build/appicon.iconset/icon_128x128@2x.png
            sips -z 256 256   build/appicon.png --out build/appicon.iconset/icon_256x256.png
            sips -z 512 512   build/appicon.png --out build/appicon.iconset/icon_256x256@2x.png
            sips -z 512 512   build/appicon.png --out build/appicon.iconset/icon_512x512.png
            sips -z 1024 1024 build/appicon.png --out build/appicon.iconset/icon_512x512@2x.png

            # Create icns file
            iconutil -c icns build/appicon.iconset -o build/appicon.icns
            echo "✅ Created appicon.icns for macOS"
        else
            echo "⚠️  'sips' command not found. Icon may not appear correctly."
            echo "   Please ensure your icon.png is at least 1024x1024 for best results."
        fi
    fi

    # For Windows, create ico file (requires ImageMagick)
    if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        echo "🪟 Detected Windows - Creating .ico file..."
        if command -v convert &> /dev/null; then
            convert build/appicon.png -define icon:auto-resize=256,128,64,48,32,16 build/appicon.ico
            echo "✅ Created appicon.ico for Windows"
        else
            echo "⚠️  ImageMagick not found. Icon may not appear correctly."
            echo "   Install with: choco install imagemagick"
        fi
    fi
else
    echo "⚠️  Warning: icon.png not found in internal/tray/"
    echo "   Using default Wails icon..."
fi

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf build/bin

# Build the application
echo "🔨 Building application..."

# Use -icon flag to specify icon path
if [[ "$OSTYPE" == "darwin"* ]] && [ -f "build/appicon.icns" ]; then
    # macOS with custom icon
    wails build -clean -icon build/appicon.icns
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] && [ -f "build/appicon.ico" ]; then
    # Windows with custom icon
    wails build -clean -icon build/appicon.ico
else
    # Default build
    wails build -clean
fi

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "✅ Build completed successfully!"
    echo ""
    echo "📦 Application built to: build/bin/"

    # Show the executable path
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "   Run with: ./build/bin/numinon-ui.app/Contents/MacOS/numinon-ui"
        echo "   Or open: open ./build/bin/numinon-ui.app"
    else
        echo "   Run with: ./build/bin/numinon-ui"
    fi
else
    echo "❌ Build failed!"
    exit 1
fi