package service

import (
	"context"
	"errors"
	"testing"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/dao"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/mock/daomock"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/pb"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Service")
}

var (
	errDAOUnknown = errors.New("unknown DAO error")
)

var _ = Describe("Service", func() {
	var (
		controller *gomock.Controller
		commentDAO *daomock.MockCommentDAO
		svc        *service
		ctx        context.Context
	)

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		commentDAO = daomock.NewMockCommentDAO(controller)
		svc = NewService(commentDAO)
		ctx = context.Background()
	})

	AfterEach(func() {
		controller.Finish()
	})

	Describe("ListComment", func() {
		var (
			req     *pb.ListCommentRequest
			videoID string
			limit   int32
			offset  int32
			resp    *pb.ListCommentResponse
			err     error
		)

		BeforeEach(func() {
			videoID = "fake id"
			limit = 10
			offset = 0
			req = &pb.ListCommentRequest{VideoId: videoID, Limit: limit, Offset: offset}
		})

		JustBeforeEach(func() {
			resp, err = svc.ListComment(ctx, req)
		})

		When("DAO error", func() {
			BeforeEach(func() {
				commentDAO.EXPECT().ListByVideoID(ctx, req.GetVideoId(), int(req.GetLimit()), int(req.GetOffset())).Return(nil, errDAOUnknown)
			})

			It("returns the error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(errDAOUnknown))
			})
		})

		When("success", func() {
			var comments []*dao.Comment

			BeforeEach(func() {
				comments = []*dao.Comment{dao.NewFakeComment(""), dao.NewFakeComment("")}
				commentDAO.EXPECT().ListByVideoID(ctx, req.GetVideoId(), int(req.GetLimit()), int(req.GetOffset())).Return(comments, nil)
			})

			It("returns comments with no error", func() {
				Expect(resp).To(Equal(&pb.ListCommentResponse{
					Comments: []*pb.CommentInfo{
						comments[0].ToProto(),
						comments[1].ToProto(),
					},
				}))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("CreateComment", func() {
		var (
			req     *pb.CreateCommentRequest
			comment *dao.Comment
			resp    *pb.CreateCommentResponse
			err     error
		)

		BeforeEach(func() {
			req = &pb.CreateCommentRequest{
				VideoId: "fake id",
				Content: "fake conetent",
			}
			comment = &dao.Comment{
				VideoID: req.GetVideoId(),
				Content: req.GetContent(),
			}
		})

		JustBeforeEach(func() {
			resp, err = svc.CreateComment(ctx, req)
		})

		When("DAO error", func() {
			BeforeEach(func() {
				commentDAO.EXPECT().Create(ctx, comment).Return(uuid.Nil, errDAOUnknown)
			})

			It("returns the error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(errDAOUnknown))
			})
		})

		When("success", func() {
			var id uuid.UUID

			BeforeEach(func() {
				id = uuid.New()
				commentDAO.EXPECT().Create(ctx, comment).Return(id, nil)
			})

			It("returns no error", func() {
				Expect(resp).To(Equal(&pb.CreateCommentResponse{
					Id: id.String(),
				}))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("UpdateComment", func() {
		var (
			req     *pb.UpdateCommentRequest
			comment *dao.Comment
			resp    *pb.UpdateCommentResponse
			err     error
		)

		BeforeEach(func() {
			req = &pb.UpdateCommentRequest{
				Id:      uuid.NewString(),
				Content: "fake content",
			}
			comment = &dao.Comment{
				ID:      uuid.MustParse(req.GetId()),
				Content: req.GetContent(),
			}
		})

		JustBeforeEach(func() {
			resp, err = svc.UpdateComment(ctx, req)
		})

		When("DAO error", func() {
			BeforeEach(func() {
				commentDAO.EXPECT().Update(ctx, comment).Return(errDAOUnknown)
			})

			It("returns the error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(errDAOUnknown))
			})
		})

		When("comment not found", func() {
			BeforeEach(func() {
				commentDAO.EXPECT().Update(ctx, comment).Return(ErrCommentNotFound)
			})

			It("return comment not found error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(ErrCommentNotFound))
			})
		})

		When("success", func() {
			BeforeEach(func() {
				commentDAO.EXPECT().Update(ctx, comment).Return(nil)
			})

			It("returns without any error", func() {
				Expect(resp).To(Equal(&pb.UpdateCommentResponse{}))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("DeleteComment", func() {
		var (
			req  *pb.DeleteCommentRequest
			resp *pb.DeleteCommentResponse
			id   uuid.UUID
			err  error
		)

		BeforeEach(func() {
			id = uuid.New()
			req = &pb.DeleteCommentRequest{Id: id.String()}
		})

		JustBeforeEach(func() {
			resp, err = svc.DeleteComment(ctx, req)
		})

		When("DAO error", func() {

			BeforeEach(func() {
				commentDAO.EXPECT().Delete(ctx, id).Return(errDAOUnknown)
			})

			It("returns the error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(errDAOUnknown))
			})
		})

		When("comment not found", func() {
			BeforeEach(func() {
				commentDAO.EXPECT().Delete(ctx, id).Return(ErrCommentNotFound)
			})

			It("return comment not found error", func() {
				Expect(resp).To(BeNil())
				Expect(err).To(MatchError(ErrCommentNotFound))
			})
		})

		When("success", func() {
			BeforeEach(func() {
				commentDAO.EXPECT().Delete(ctx, id).Return(nil)
			})

			It("returns without any error", func() {
				Expect(resp).To(Equal(&pb.DeleteCommentResponse{}))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
