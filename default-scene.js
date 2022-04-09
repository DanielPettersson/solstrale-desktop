green = solidLambertian(.5, 1, .5)
snow = solidLambertian(1, 1, 1)
metalMat = solidMetal(.6, .6, .6, .5)
glass = solidDielectric(1, 1, .7, 1.5)

function createSnowman(offset, mat, angle) {
    black = solidLambertian(.1, .1, .1)
    orange = solidLambertian(1, .4, .3)
    redLight = light(2, 0, 0)

    parts = []
    parts.push(sphere(vec(0,.9,0), 1, mat))
    parts.push(sphere(vec(0,2.5,0), .7, mat))
    parts.push(sphere(vec(0,3.5,0), .4, mat))
    parts.push(box(vec(-1.5, 2.5, 0), vec(1.5, 2.6, .1), black))
    parts.push(box(vec(-.05, 3.5, 0), vec(.05, 3.6,  1), orange))
    parts.push(sphere(vec(0,2.7, .7), .06, black))
    parts.push(sphere(vec(0,2.5, .73), .06, black))
    parts.push(sphere(vec(0,2.3, .7), .06, black))
    parts.push(sphere(vec(-.2,3.65, .35), .08, redLight))
    parts.push(sphere(vec(.2,3.65, .35), .08, redLight))

    bvh = boundingVolumeHierarchy(parts)
	rotated = rotationY(bvh, angle)
    return translation(rotated, offset) 
}

scene = scene(
    hittableList([
        sphere(vec(5,5,10), 3, light(15, 15, 15)),
        sphere(vec(3,8,8), 3, light(5, 5, 5)),
        quad(vec(-100, 0, -100), vec(200, 0, 0), vec(0, 0, 200), green),
		createSnowman(vec(0, 0, 0), snow, 0),
		createSnowman(vec(-3, 0, 1), metalMat, 30),
		createSnowman(vec(3, 0, 1), glass, -45)
    ]),
    camera(40, vec(-3, 4, 10), vec(0, 1.5, 0), .5, 10),
    col(0, 0, 0),
    renderConfig(1000)
)
