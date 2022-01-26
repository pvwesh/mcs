package main

//필요한 패키지 import
import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/geo/s2"
)

//geojson 데이터 형태에 맞게 struct 구성 및 정의
type LinkStr struct {
	Type     string
	Features []struct {
		Type     string
		Geometry struct {
			Type        string
			Coordinates [][]float64
		}
		Properties struct {
			ID string
		}
	}
}

func main() {
	// CLI 프로그램에 링크파일 이름을 -links, 타겟의 위도를 -targetLat, 타겟의 경도를 -targetLng로 입력할 수 있도록 설정
	// 기본값으로 과제의 예제에 표현된 값 설정
	linksflg := flag.String("links", "links.geojson", "a geojson file name")
	targetLatflg := flag.Float64("targetLat", 37.499212063, "target Latitutude float64")
	targetLngflg := flag.Float64("targetLng", 127.027268062, "target Longitude float64")
	flag.Parse()

	// 링크파일 열기
	data, _ := os.Open(*linksflg)

	// open된 데이터를 바이트 데이터로 읽기
	byteValue, _ := ioutil.ReadAll(data)

	// 읽은 링크파일을 struct화 시킬 수 있도록 변수 선언
	var linksDB LinkStr

	// 바이트 데이터를 struct에 저장
	json.Unmarshal(byteValue, &linksDB)

	// S2 라이브러리에서 함수를 사용하기 위해 json데이터의 feature를 polyline 형태로 변환
	// for문을 활용하여 struct에 저장된 feature의 좌표를 기반으로 포인트, 폴리라인 순으로 변환함

	// 모든 링크를 담는 polylines 리스트 변수 설정
	polylines := []s2.Polyline{}
	// s2.Polyline: 포함된 포인트들을 기반으로 폴리라인을 생성하는 함수
	poly := s2.Polyline{}
	for _, i := range linksDB.Features {
		for _, j := range i.Geometry.Coordinates {
			// s2.LatLngFromDegrees : struct에 존재하는 위경도 데이터를 s2 라이브러리가 가정하는 3차원 구 위경도(s2 LatLng)로 변환
			// s2.PointFromLatLng : s2 LatLng 위경도를 활용하여 포인트 생성
			// append를 통해 생성된 포인트들이 순차적으로 s2.Polyline함수에 포함되도록 함
			poly = append(poly, s2.PointFromLatLng(s2.LatLngFromDegrees(j[1], j[0])))
		}
		// 각각의 feature가 polyline으로 변환된 뒤 polylines 리스트에 순차적으로 더해져 polylines 데이터셋 생성
		polylines = append(polylines, poly)
	}

	// 최근접 경로 쿼리에 활용하기 위하여 polyline과 edge Index 부여
	index := s2.NewShapeIndex()
	for _, l := range polylines {
		index.Add(&l)
	}

	// 타겟 좌표의 위경도 입력값을 활용하여 타겟 좌표에 해당하는 포인트 생성
	targetpoint := s2.PointFromLatLng(s2.LatLngFromDegrees(*targetLatflg, *targetLngflg))

	// 생성된 타겟 포인트를 최근접 경로 쿼리 타겟으로 설정
	target := s2.NewMinDistanceToPointTarget(targetpoint)

	// 최근접 경로를 찾기 위한 쿼리 옵션 설정
	opts := s2.NewClosestEdgeQueryOptions()
	opts.MaxResults(1)

	// 최근접 경로 찾기 쿼리 선언
	query := s2.NewClosestEdgeQuery(index, opts)

	// 타겟을 기반으로 쿼리 수행
	for _, result := range query.FindEdges(target) {
		// 쿼리로 찾은 인덱스를 기반으로 최근접 엣지 추출
		ClosestEdge := index.Shape(result.ShapeID()).Edge(int(result.EdgeID()))
		// 최근접 엣지까지의 최근접 거리계산
		ClosestDistance := result.Distance().Angle().Radians() * 6371010 //NASA에서 제공하는 지구의 평균 반지름 미터값
		// 최근접 엣지 위 최근접 좌표추출
		ClosestPoint := s2.LatLngFromPoint(s2.Project(targetpoint, ClosestEdge.V0, ClosestEdge.V1))
		// 결과 표츌
		fmt.Println(ClosestDistance, ", ", ClosestPoint.Lng.Degrees(), ", ", ClosestPoint.Lat.Degrees())

	}
	// CLI 프로그램에 표출된 결과값 확인 후 Enter를 누르면 종료되도록 설정
	fmt.Scanln()
}
