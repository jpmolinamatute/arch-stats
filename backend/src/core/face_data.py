from schema.face_schema import Face, FaceType, Ring, Spot

face_data: list[Face] = [
    Face(
        face_type=FaceType.NONE,
        face_name="No Target Face",
        render_cross=False,
        viewBox=0.0,
        spots=[],
        rings=[],
    ),
    Face(
        face_name="WA 122cm Standard Target Face",
        face_type=FaceType.WA_122_FULL,
        render_cross=True,
        viewBox=1342.0,
        spots=[
            Spot(
                x_offset=0.0,
                y_offset=0.0,
                diameter=1220.0,
            )
        ],
        rings=[
            Ring(
                data_score=1,
                fill="#FFFFFF",
                r=610.0,  # 1220mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=2,
                fill="#FFFFFF",
                r=549.0,  # 1098mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=3,
                fill="#000000",
                r=488.0,  # 976mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=4,
                fill="#000000",
                r=427.0,  # 854mm / 2
                stroke="#FFFFFF",
                stroke_width=1.0,
            ),
            Ring(
                data_score=5,
                fill="#00B4E4",
                r=366.0,  # 732mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=6,
                fill="#00B4E4",
                r=305.0,  # 610mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=7,
                fill="#F65058",
                r=244.0,  # 488mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=8,
                fill="#F65058",
                r=183.0,  # 366mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=9,
                fill="#FFE552",
                r=122.0,  # 244mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=10,
                fill="#FFE552",
                r=61.0,  # 122mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=10,
                fill="#FFE552",
                r=30.5,  # 61mm diameter / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
        ],
    ),
    Face(
        face_name="WA 80cm Standard Target Face",
        face_type=FaceType.WA_80_FULL,
        render_cross=True,
        viewBox=880.0,
        spots=[
            Spot(
                x_offset=0.0,
                y_offset=0.0,
                diameter=800.0,
            )
        ],
        rings=[
            Ring(
                data_score=1,
                fill="#FFFFFF",
                r=400.0,  # 800mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=2,
                fill="#FFFFFF",
                r=360.0,  # 720mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=3,
                fill="#000000",
                r=320.0,  # 640mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=4,
                fill="#000000",
                r=280.0,  # 560mm / 2
                stroke="#FFFFFF",
                stroke_width=1.0,
            ),
            Ring(
                data_score=5,
                fill="#00B4E4",
                r=240.0,  # 480mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=6,
                fill="#00B4E4",
                r=200.0,  # 400mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=7,
                fill="#F65058",
                r=160.0,  # 320mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=8,
                fill="#F65058",
                r=120.0,  # 240mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=9,
                fill="#FFE552",
                r=80.0,  # 160mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=10,
                fill="#FFE552",
                r=40.0,  # 80mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=10,
                fill="#FFE552",
                r=20.0,  # 40mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
        ],
    ),
    Face(
        face_name="WA 60cm Standard Target Face",
        face_type=FaceType.WA_60_FULL,
        render_cross=True,
        viewBox=660.0,
        spots=[
            Spot(
                x_offset=0.0,
                y_offset=0.0,
                diameter=600.0,
            )
        ],
        rings=[
            Ring(
                data_score=1,
                fill="#FFFFFF",
                r=300.0,  # 600mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=2,
                fill="#FFFFFF",
                r=270.0,  # 540mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=3,
                fill="#000000",
                r=240.0,  # 480mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=4,
                fill="#000000",
                r=210.0,  # 420mm / 2
                stroke="#FFFFFF",
                stroke_width=1.0,
            ),
            Ring(
                data_score=5,
                fill="#00B4E4",
                r=180.0,  # 360mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=6,
                fill="#00B4E4",
                r=150.0,  # 300mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=7,
                fill="#F65058",
                r=120.0,  # 240mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=8,
                fill="#F65058",
                r=90.0,  # 180mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=9,
                fill="#FFE552",
                r=60.0,  # 120mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=10,
                fill="#FFE552",
                r=30.0,  # 60mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=10,
                fill="#FFE552",
                r=15.0,  # 30mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
        ],
    ),
    Face(
        face_name="WA 40cm Standard Target Face",
        face_type=FaceType.WA_40_FULL,
        render_cross=True,
        viewBox=440.0,
        spots=[
            Spot(
                x_offset=0.0,
                y_offset=0.0,
                diameter=400.0,
            )
        ],
        rings=[
            Ring(
                data_score=1,
                fill="#FFFFFF",
                r=200.0,  # 400mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=2,
                fill="#FFFFFF",
                r=180.0,  # 360mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=3,
                fill="#000000",
                r=160.0,  # 320mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=4,
                fill="#000000",
                r=140.0,  # 280mm / 2
                stroke="#FFFFFF",
                stroke_width=1.0,
            ),
            Ring(
                data_score=5,
                fill="#00B4E4",
                r=120.0,  # 240mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=6,
                fill="#00B4E4",
                r=100.0,  # 200mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=7,
                fill="#F65058",
                r=80.0,  # 160mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=8,
                fill="#F65058",
                r=60.0,  # 120mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=9,
                fill="#FFE552",
                r=40.0,  # 80mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=10,
                fill="#FFE552",
                r=20.0,  # 40mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=10,
                fill="#FFE552",
                r=10.0,  # 20mm / 2 (recurve X)
                stroke="#000000",
                stroke_width=1.0,
            ),
        ],
    ),
    Face(
        face_name="WA 122cm 6-Ring Target Face",
        face_type=FaceType.WA_122_6RINGS,
        render_cross=True,
        viewBox=854.0,
        spots=[
            Spot(
                x_offset=0.0,
                y_offset=0.0,
                diameter=732.0,  # 5-ring outer diameter (366mm radius * 2)
            )
        ],
        rings=[
            Ring(
                data_score=5,
                fill="#00B4E4",
                r=366.0,  # 732mm / 2 (outermost for 6-ring)
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=6,
                fill="#00B4E4",
                r=305.0,  # 610mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=7,
                fill="#F65058",
                r=244.0,  # 488mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=8,
                fill="#F65058",
                r=183.0,  # 366mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=9,
                fill="#FFE552",
                r=122.0,  # 244mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=10,
                fill="#FFE552",
                r=61.0,  # 122mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=10,
                fill="#FFE552",
                r=30.5,  # 61mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
        ],
    ),
    Face(
        face_name="WA 80cm 6-Ring Target Face",
        face_type=FaceType.WA_80_6RINGS,
        render_cross=True,
        viewBox=560.0,
        spots=[
            Spot(
                x_offset=0.0,
                y_offset=0.0,
                diameter=480.0,  # 5-ring outer diameter (240mm radius * 2)
            )
        ],
        rings=[
            Ring(
                data_score=5,
                fill="#00B4E4",
                r=240.0,  # 480mm / 2 (outermost for 6-ring)
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=6,
                fill="#00B4E4",
                r=200.0,  # 400mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=7,
                fill="#F65058",
                r=160.0,  # 320mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=8,
                fill="#F65058",
                r=120.0,  # 240mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=9,
                fill="#FFE552",
                r=80.0,  # 160mm / 2
                stroke="#000000",
                stroke_width=2.0,
            ),
            Ring(
                data_score=10,
                fill="#FFE552",
                r=40.0,  # 80mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
            Ring(
                data_score=10,
                fill="#FFE552",
                r=20.0,  # 40mm / 2
                stroke="#000000",
                stroke_width=1.0,
            ),
        ],
    ),
]
