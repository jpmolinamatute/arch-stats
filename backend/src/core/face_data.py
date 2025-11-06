from schema.face_schema import Face, FaceType, Ring, Spot


face_data: list[Face] = [
    Face(
        face_id=FaceType.NONE,
        face_name="No Target Face",
        svg_scale_factor=1.0,
        render_cross=False,
        spots=[],
        rings=[],
    ),
    Face(
        face_name="WA 122cm Standard Target Face",
        face_id=FaceType.WA_122_FULL,
        svg_scale_factor=1.0,
        render_cross=True,
        spots=[
            Spot(
                x_offset=0.0,
                y_offset=0.0,
                diameter=1220.0,
            )
        ],
        rings=[
            # X-ring (inner 10)
            Ring(
                score=10,
                scoring_zone_label="X",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=30.5,  # 61mm diameter / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            # 10 through 1 rings (outer radius in mm)
            Ring(
                score=10,
                scoring_zone_label="10",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=61.0,  # 122mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=9,
                scoring_zone_label="9",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=122.0,  # 244mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=8,
                scoring_zone_label="8",
                scoring_zone_color="#F65058",
                scoring_zone_radius=183.0,  # 366mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=7,
                scoring_zone_label="7",
                scoring_zone_color="#F65058",
                scoring_zone_radius=244.0,  # 488mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=6,
                scoring_zone_label="6",
                scoring_zone_color="#00B4E4",
                scoring_zone_radius=305.0,  # 610mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=5,
                scoring_zone_label="5",
                scoring_zone_color="#00B4E4",
                scoring_zone_radius=366.0,  # 732mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=4,
                scoring_zone_label="4",
                scoring_zone_color="#000000",
                scoring_zone_radius=427.0,  # 854mm / 2
                outer_line_color="#FFFFFF",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=3,
                scoring_zone_label="3",
                scoring_zone_color="#000000",
                scoring_zone_radius=488.0,  # 976mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=2,
                scoring_zone_label="2",
                scoring_zone_color="#FFFFFF",
                scoring_zone_radius=549.0,  # 1098mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=1,
                scoring_zone_label="1",
                scoring_zone_color="#FFFFFF",
                scoring_zone_radius=610.0,  # 1220mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
        ],
    ),
    Face(
        face_name="WA 80cm Standard Target Face",
        face_id=FaceType.WA_80_FULL,
        svg_scale_factor=1.0,
        render_cross=True,
        spots=[
            Spot(
                x_offset=0.0,
                y_offset=0.0,
                diameter=800.0,
            )
        ],
        rings=[
            Ring(
                score=10,
                scoring_zone_label="X",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=20.0,  # 40mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=10,
                scoring_zone_label="10",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=40.0,  # 80mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=9,
                scoring_zone_label="9",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=80.0,  # 160mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=8,
                scoring_zone_label="8",
                scoring_zone_color="#F65058",
                scoring_zone_radius=120.0,  # 240mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=7,
                scoring_zone_label="7",
                scoring_zone_color="#F65058",
                scoring_zone_radius=160.0,  # 320mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=6,
                scoring_zone_label="6",
                scoring_zone_color="#00B4E4",
                scoring_zone_radius=200.0,  # 400mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=5,
                scoring_zone_label="5",
                scoring_zone_color="#00B4E4",
                scoring_zone_radius=240.0,  # 480mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=4,
                scoring_zone_label="4",
                scoring_zone_color="#000000",
                scoring_zone_radius=280.0,  # 560mm / 2
                outer_line_color="#FFFFFF",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=3,
                scoring_zone_label="3",
                scoring_zone_color="#000000",
                scoring_zone_radius=320.0,  # 640mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=2,
                scoring_zone_label="2",
                scoring_zone_color="#FFFFFF",
                scoring_zone_radius=360.0,  # 720mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=1,
                scoring_zone_label="1",
                scoring_zone_color="#FFFFFF",
                scoring_zone_radius=400.0,  # 800mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
        ],
    ),
    Face(
        face_name="WA 60cm Standard Target Face",
        face_id=FaceType.WA_60_FULL,
        svg_scale_factor=1.0,
        render_cross=True,
        spots=[
            Spot(
                x_offset=0.0,
                y_offset=0.0,
                diameter=600.0,
            )
        ],
        rings=[
            Ring(
                score=10,
                scoring_zone_label="X",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=15.0,  # 30mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=10,
                scoring_zone_label="10",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=30.0,  # 60mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=9,
                scoring_zone_label="9",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=60.0,  # 120mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=8,
                scoring_zone_label="8",
                scoring_zone_color="#F65058",
                scoring_zone_radius=90.0,  # 180mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=7,
                scoring_zone_label="7",
                scoring_zone_color="#F65058",
                scoring_zone_radius=120.0,  # 240mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=6,
                scoring_zone_label="6",
                scoring_zone_color="#00B4E4",
                scoring_zone_radius=150.0,  # 300mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=5,
                scoring_zone_label="5",
                scoring_zone_color="#00B4E4",
                scoring_zone_radius=180.0,  # 360mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=4,
                scoring_zone_label="4",
                scoring_zone_color="#000000",
                scoring_zone_radius=210.0,  # 420mm / 2
                outer_line_color="#FFFFFF",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=3,
                scoring_zone_label="3",
                scoring_zone_color="#000000",
                scoring_zone_radius=240.0,  # 480mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=2,
                scoring_zone_label="2",
                scoring_zone_color="#FFFFFF",
                scoring_zone_radius=270.0,  # 540mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=1,
                scoring_zone_label="1",
                scoring_zone_color="#FFFFFF",
                scoring_zone_radius=300.0,  # 600mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
        ],
    ),
    Face(
        face_name="WA 40cm Standard Target Face",
        face_id=FaceType.WA_40_FULL,
        svg_scale_factor=1.0,
        render_cross=True,
        spots=[
            Spot(
                x_offset=0.0,
                y_offset=0.0,
                diameter=400.0,
            )
        ],
        rings=[
            Ring(
                score=10,
                scoring_zone_label="X",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=10.0,  # 20mm / 2 (recurve X)
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=10,
                scoring_zone_label="10",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=20.0,  # 40mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=9,
                scoring_zone_label="9",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=40.0,  # 80mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=8,
                scoring_zone_label="8",
                scoring_zone_color="#F65058",
                scoring_zone_radius=60.0,  # 120mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=7,
                scoring_zone_label="7",
                scoring_zone_color="#F65058",
                scoring_zone_radius=80.0,  # 160mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=6,
                scoring_zone_label="6",
                scoring_zone_color="#00B4E4",
                scoring_zone_radius=100.0,  # 200mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=5,
                scoring_zone_label="5",
                scoring_zone_color="#00B4E4",
                scoring_zone_radius=120.0,  # 240mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=4,
                scoring_zone_label="4",
                scoring_zone_color="#000000",
                scoring_zone_radius=140.0,  # 280mm / 2
                outer_line_color="#FFFFFF",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=3,
                scoring_zone_label="3",
                scoring_zone_color="#000000",
                scoring_zone_radius=160.0,  # 320mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=2,
                scoring_zone_label="2",
                scoring_zone_color="#FFFFFF",
                scoring_zone_radius=180.0,  # 360mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=1,
                scoring_zone_label="1",
                scoring_zone_color="#FFFFFF",
                scoring_zone_radius=200.0,  # 400mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
        ],
    ),
    Face(
        face_name="WA 122cm 6-Ring Target Face",
        face_id=FaceType.WA_122_6RINGS,
        svg_scale_factor=1.0,
        render_cross=True,
        spots=[
            Spot(
                x_offset=0.0,
                y_offset=0.0,
                diameter=732.0,  # 5-ring outer diameter (366mm radius * 2)
            )
        ],
        rings=[
            # X-ring (inner 10)
            Ring(
                score=10,
                scoring_zone_label="X",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=30.5,  # 61mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=10,
                scoring_zone_label="10",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=61.0,  # 122mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=9,
                scoring_zone_label="9",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=122.0,  # 244mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=8,
                scoring_zone_label="8",
                scoring_zone_color="#F65058",
                scoring_zone_radius=183.0,  # 366mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=7,
                scoring_zone_label="7",
                scoring_zone_color="#F65058",
                scoring_zone_radius=244.0,  # 488mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=6,
                scoring_zone_label="6",
                scoring_zone_color="#00B4E4",
                scoring_zone_radius=305.0,  # 610mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=5,
                scoring_zone_label="5",
                scoring_zone_color="#00B4E4",
                scoring_zone_radius=366.0,  # 732mm / 2 (outermost for 6-ring)
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
        ],
    ),
    Face(
        face_name="WA 80cm 6-Ring Target Face",
        face_id=FaceType.WA_80_6RINGS,
        svg_scale_factor=1.0,
        render_cross=True,
        spots=[
            Spot(
                x_offset=0.0,
                y_offset=0.0,
                diameter=480.0,  # 5-ring outer diameter (240mm radius * 2)
            )
        ],
        rings=[
            Ring(
                score=10,
                scoring_zone_label="X",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=20.0,  # 40mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=10,
                scoring_zone_label="10",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=40.0,  # 80mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=9,
                scoring_zone_label="9",
                scoring_zone_color="#FFE552",
                scoring_zone_radius=80.0,  # 160mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=8,
                scoring_zone_label="8",
                scoring_zone_color="#F65058",
                scoring_zone_radius=120.0,  # 240mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=7,
                scoring_zone_label="7",
                scoring_zone_color="#F65058",
                scoring_zone_radius=160.0,  # 320mm / 2
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
            Ring(
                score=6,
                scoring_zone_label="6",
                scoring_zone_color="#00B4E4",
                scoring_zone_radius=200.0,  # 400mm / 2
                outer_line_color="#000000",
                outer_line_thickness=1.0,
            ),
            Ring(
                score=5,
                scoring_zone_label="5",
                scoring_zone_color="#00B4E4",
                scoring_zone_radius=240.0,  # 480mm / 2 (outermost for 6-ring)
                outer_line_color="#000000",
                outer_line_thickness=2.0,
            ),
        ],
    ),
]
