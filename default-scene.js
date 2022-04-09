scene = scene(
    hittableList([
        sphere(vec(0,0,0), 1, solidLambertian(1, 1, 1)),
        sphere(vec(5,10,0), 5, light(10, 10, 10))
    ]),
    simpleCamera(40, vec(0, 0, 10), vec(0, 0, 0)),
    col(1, 0, 0),
    renderConfig(100)
)
