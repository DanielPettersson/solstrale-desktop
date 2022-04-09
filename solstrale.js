function vec(x, y, z) {
    return { x: x, y: y, z: z }
}

function col(r, g, b) {
    return { r: r, g: g, b: b }
}

function simpleCamera(verticalFovDegrees, lookFrom, lookAt) {
    return {
        verticalFovDegrees: verticalFovDegrees,
        apertureSize: 0,
        focusDistance: 1,
        lookFrom: lookFrom,
        lookAt: lookAt,
        vup: vec(0, 1, 0)
    }
}

function renderConfig(samplesPerPixel) {
    return {
        imageWidth: windowWidth,
        imageHeight: windowHeight,
        samplesPerPixel: samplesPerPixel,
        shader: {
            type: "pathTracing",
            maxDepth: 50
        }
    }
}

function hittableList(items) {
    return {
        type: "hittableList",
        list: items
    }
}

function sphere(center, radius, mat) {
    return {
        type: "sphere",
        center: center,
        radius: radius,
        mat: mat
    }
}

function solidLambertian(r, g, b) {
    return {
        type: "lambertian",
        texture: solidColor(col(r, g, b))
    }
}

function light(r, g, b) {
    return {
        type: "diffuseLight",
        color: col(r, g, b)
    }
}

function solidColor(col) {
    return {
        type: "solidColor",
        color: col
    }
}

function scene(world, camera, background, renderConfig) {
    return {
        world: world,
        camera: camera,
        background: background,
        renderConfig: renderConfig
    }
}