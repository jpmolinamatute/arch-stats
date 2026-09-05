package repository

import "github.com/jpmolinamatute/arch-stats/backend/internal/model"

// DefaultFaceCatalog contains the standardized World Archery target face definitions
// ported from face_data.py (14KB Python source).
var DefaultFaceCatalog = []model.FaceRead{
	{
		FaceType:    model.FaceTypeNone,
		FaceName:    "No Target Face",
		RenderCross: false,
		ViewBox:     0.0,
		Spots:       []model.Spot{},
		Rings:       []model.Ring{},
	},
	{
		FaceType:    model.FaceTypeWA122Full,
		FaceName:    "WA 122cm Standard Target Face",
		RenderCross: true,
		ViewBox:     1342.0,
		Spots: []model.Spot{
			{
				XOffset:  0.0,
				YOffset:  0.0,
				Diameter: 1220.0,
			},
		},
		Rings: []model.Ring{
			{DataScore: 1, Fill: "#FFFFFF", R: 610.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 2, Fill: "#FFFFFF", R: 549.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 3, Fill: "#000000", R: 488.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 4, Fill: "#000000", R: 427.0, Stroke: "#FFFFFF", StrokeWidth: 1.0},
			{DataScore: 5, Fill: "#00B4E4", R: 366.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 6, Fill: "#00B4E4", R: 305.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 7, Fill: "#F65058", R: 244.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 8, Fill: "#F65058", R: 183.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 9, Fill: "#FFE552", R: 122.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 10, Fill: "#FFE552", R: 61.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 10, Fill: "#FFE552", R: 30.5, Stroke: "#000000", StrokeWidth: 1.0},
		},
	},
	{
		FaceType:    model.FaceTypeWA80Full,
		FaceName:    "WA 80cm Standard Target Face",
		RenderCross: true,
		ViewBox:     880.0,
		Spots: []model.Spot{
			{
				XOffset:  0.0,
				YOffset:  0.0,
				Diameter: 800.0,
			},
		},
		Rings: []model.Ring{
			{DataScore: 1, Fill: "#FFFFFF", R: 400.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 2, Fill: "#FFFFFF", R: 360.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 3, Fill: "#000000", R: 320.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 4, Fill: "#000000", R: 280.0, Stroke: "#FFFFFF", StrokeWidth: 1.0},
			{DataScore: 5, Fill: "#00B4E4", R: 240.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 6, Fill: "#00B4E4", R: 200.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 7, Fill: "#F65058", R: 160.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 8, Fill: "#F65058", R: 120.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 9, Fill: "#FFE552", R: 80.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 10, Fill: "#FFE552", R: 40.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 10, Fill: "#FFE552", R: 20.0, Stroke: "#000000", StrokeWidth: 1.0},
		},
	},
	{
		FaceType:    model.FaceTypeWA60Full,
		FaceName:    "WA 60cm Standard Target Face",
		RenderCross: true,
		ViewBox:     660.0,
		Spots: []model.Spot{
			{
				XOffset:  0.0,
				YOffset:  0.0,
				Diameter: 600.0,
			},
		},
		Rings: []model.Ring{
			{DataScore: 1, Fill: "#FFFFFF", R: 300.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 2, Fill: "#FFFFFF", R: 270.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 3, Fill: "#000000", R: 240.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 4, Fill: "#000000", R: 210.0, Stroke: "#FFFFFF", StrokeWidth: 1.0},
			{DataScore: 5, Fill: "#00B4E4", R: 180.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 6, Fill: "#00B4E4", R: 150.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 7, Fill: "#F65058", R: 120.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 8, Fill: "#F65058", R: 90.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 9, Fill: "#FFE552", R: 60.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 10, Fill: "#FFE552", R: 30.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 10, Fill: "#FFE552", R: 15.0, Stroke: "#000000", StrokeWidth: 1.0},
		},
	},
	{
		FaceType:    model.FaceTypeWA40Full,
		FaceName:    "WA 40cm Standard Target Face",
		RenderCross: true,
		ViewBox:     440.0,
		Spots: []model.Spot{
			{
				XOffset:  0.0,
				YOffset:  0.0,
				Diameter: 400.0,
			},
		},
		Rings: []model.Ring{
			{DataScore: 1, Fill: "#FFFFFF", R: 200.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 2, Fill: "#FFFFFF", R: 180.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 3, Fill: "#000000", R: 160.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 4, Fill: "#000000", R: 140.0, Stroke: "#FFFFFF", StrokeWidth: 1.0},
			{DataScore: 5, Fill: "#00B4E4", R: 120.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 6, Fill: "#00B4E4", R: 100.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 7, Fill: "#F65058", R: 80.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 8, Fill: "#F65058", R: 60.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 9, Fill: "#FFE552", R: 40.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 10, Fill: "#FFE552", R: 20.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 10, Fill: "#FFE552", R: 10.0, Stroke: "#000000", StrokeWidth: 1.0},
		},
	},
	{
		FaceType:    model.FaceTypeWA1226Rings,
		FaceName:    "WA 122cm 6-Ring Target Face",
		RenderCross: true,
		ViewBox:     854.0,
		Spots: []model.Spot{
			{
				XOffset:  0.0,
				YOffset:  0.0,
				Diameter: 732.0,
			},
		},
		Rings: []model.Ring{
			{DataScore: 5, Fill: "#00B4E4", R: 366.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 6, Fill: "#00B4E4", R: 305.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 7, Fill: "#F65058", R: 244.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 8, Fill: "#F65058", R: 183.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 9, Fill: "#FFE552", R: 122.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 10, Fill: "#FFE552", R: 61.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 10, Fill: "#FFE552", R: 30.5, Stroke: "#000000", StrokeWidth: 1.0},
		},
	},
	{
		FaceType:    model.FaceTypeWA806Rings,
		FaceName:    "WA 80cm 6-Ring Target Face",
		RenderCross: true,
		ViewBox:     560.0,
		Spots: []model.Spot{
			{
				XOffset:  0.0,
				YOffset:  0.0,
				Diameter: 480.0,
			},
		},
		Rings: []model.Ring{
			{DataScore: 5, Fill: "#00B4E4", R: 240.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 6, Fill: "#00B4E4", R: 200.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 7, Fill: "#F65058", R: 160.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 8, Fill: "#F65058", R: 120.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 9, Fill: "#FFE552", R: 80.0, Stroke: "#000000", StrokeWidth: 2.0},
			{DataScore: 10, Fill: "#FFE552", R: 40.0, Stroke: "#000000", StrokeWidth: 1.0},
			{DataScore: 10, Fill: "#FFE552", R: 20.0, Stroke: "#000000", StrokeWidth: 1.0},
		},
	},
}
