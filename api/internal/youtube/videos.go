package youtube

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pkulik0/autocc/api/internal/pb"
	"github.com/pkulik0/autocc/api/internal/quota"
)

const (
	// Maximum value according to the API documentation.
	videosMaxResults = 50
)

func (y *youtube) GetVideos(ctx context.Context, userID, nextPageToken string) ([]*pb.Video, string, error) {
	if userID == "" {
		return nil, "", ErrInvalidInput
	}

	service, err := y.getInstance(ctx, userID, quota.YoutubeSearchList)
	if err != nil {
		return nil, "", err
	}

	call := service.Search.List([]string{"snippet"}).ForMine(true).MaxResults(videosMaxResults).Type("video")
	if nextPageToken != "" {
		call.PageToken(nextPageToken)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, "", err
	}

	var videos []*pb.Video
	for _, item := range resp.Items {
		if item.Id.Kind != "youtube#video" {
			continue
		}

		publishedAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			continue
		}

		videos = append(videos, &pb.Video{
			Id:           item.Id.VideoId,
			Title:        item.Snippet.Title,
			ThumbnailUrl: item.Snippet.Thumbnails.High.Url,
			Description:  item.Snippet.Description,
			PublishedAt:  &timestamppb.Timestamp{Seconds: publishedAt.Unix()},
		})
	}

	return videos, resp.NextPageToken, nil
}
