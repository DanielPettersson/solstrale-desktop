scene = scene(
    hittableList([
        sphere(vec(10,30,3), 10, light(14, 14, 13)),
        sphere(vec(0,0,0), 5, lambertian(solidColor(col(1,0,0)))),
    ]),    
    camera(50, vec(-10.5, 2, 1.5), vec(0, 3, 0), 0.1, 15),
    col(.1, .1, .2),
    renderConfig(200, oidnPostProcessor('/home/daniel/oidnDenoise'))
)