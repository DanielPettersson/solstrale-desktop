// General stuff

function vec(x, y, z) {
    return { x: x, y: y, z: z }
}

function col(r, g, b) {
    return { r: r, g: g, b: b }
}

function scene(world, camera, background, renderConfig) {
    return {
        world: world,
        camera: camera,
        background: background,
        renderConfig: renderConfig
    }
}

function camera(verticalFovDegrees, lookFrom, lookAt, apertureSize, focusDistance) {
    return {
        verticalFovDegrees: verticalFovDegrees,
        apertureSize: apertureSize,
        focusDistance: focusDistance,
        lookFrom: lookFrom,
        lookAt: lookAt,
        vup: vec(0, 1, 0)
    }
}

function renderConfig(samplesPerPixel, postProcessor) {
    return {
        imageWidth: windowWidth,
        imageHeight: windowHeight,
        samplesPerPixel: samplesPerPixel,
        shader: {
            type: "pathTracing",
            maxDepth: 50
        },
        postProcessor: postProcessor
    }
}

function oidnPostProcessor(oidnDenoiseExecutablePath) {
    return {
        type: "oidn",
        oidnDenoiseExecutablePath: oidnDenoiseExecutablePath
    }
}

// hittables

function hittableList(items) {
    return {
        type: "hittableList",
        list: items
    }
}

function boundingVolumeHierarchy(items) {
    return {
        type: "bvh",
        list: items
    }
}

function constantMedium(object, density, col) {
    return {
        type: "constantMedium",
        object: object,
        density: density,
        texture: solidColor(col)
    }
}

function motionBlur(object, blurDirection) {
    return {
        type: "motionBlur",
        object: object,
        blurDirection: blurDirection
    }
}

function rotationY(object, angle) {
    return {
        type: "rotationY",
        object: object,
        angle: angle
    }
}

function translation(object, offset) {
    return {
        type: "translation",
        object: object,
        offset: offset
    }
}

function box(corner, diagonalCorner, mat) {
    return {
        type: "box",
        corner: corner,
        diagonalCorner: diagonalCorner,
        mat: mat
    }
}

function quad(corner, dirU, dirV, mat) {
    return {
        type: "quad",
        corner: corner,
        dirU: dirU,
        dirV: dirV,
        mat: mat
    }
}

function triangle(v0, v1, v2, mat) {
    return {
        type: "triangle",
        v0: v0,
        v1: v1,
        v2: v2,
        mat: mat
    }
}

function objModel(path) {
    return {
        type: "objModel",
        path: path
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

// materials

function solidLambertian(r, g, b) {
    return lambertian(solidColor(col(r, g, b)))
}

function lambertian(tex) {
    return {
        type: "lambertian",
        texture: tex
    }
}

function solidMetal(r, g, b, fuzz) {
    return metal(solidColor(col(r, g, b)), fuzz)
}

function metal(tex, fuzz) {
    return {
        type: "metal",
        texture: tex,
        fuzz: fuzz
    }
}

function solidDielectric(r, g, b, indexOfRefraction) {
    return dielectric(solidColor(col(r, g, b)), indexOfRefraction)
}

function dielectric(tex, indexOfRefraction) {
    return {
        type: "dielectric",
        texture: tex,
        indexOfRefraction: indexOfRefraction
    }
}

function light(r, g, b) {
    return {
        type: "diffuseLight",
        color: col(r, g, b)
    }
}

// textures

function solidColor(col) {
    return {
        type: "solidColor",
        color: col
    }
}

function checker(scale, even, odd) {
    return {
        type: "checker",
        scale: scale,
        even: even,
        odd: odd,
    }
}

function image(path) {
    return {
        type: "image",
        path: path,
        mirror: false
    }
}

function noise(scale, col) {
    return {
        type: "noise",
        scale: scale,
        col: col
    }
}
