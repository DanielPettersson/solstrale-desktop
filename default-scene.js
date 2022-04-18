scene = scene(
    hittableList([
        sphere(vec(10,30,0), 10, light(10, 10, 10)),
        sphere(vec(11.2,6.5,-1.4), .7, dielectric(solidColor(col(1, .6, .6)), 1.5)),
        sphere(vec(11.2,6.5, 1.3), .7, dielectric(solidColor(col(.6, 1, .6)), 1.5)),
        sphere(vec(0, 6, 5), 1.5, metal(solidColor(col(1, 1, .6)), .2)),
        objModel('sponza.obj')
    ]),    
    camera(50, vec(15, 7.5, 5), vec(0, 5, -5), 0, 1),
    col(0, 0, 0),
    renderConfig(100, oidnPostProcessor('/home/daniel/oidnDenoise'))
)