package controller

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/api"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/helper"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/log"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/template"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/validate"
	"github.com/ic3network/mccs-alpha-api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type adminEntityHandler struct {
	once *sync.Once
}

var AdminEntityHandler = newAdminEntityHandler()

func newAdminEntityHandler() *adminEntityHandler {
	return &adminEntityHandler{
		once: new(sync.Once),
	}
}

func (handler *adminEntityHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	handler.once.Do(func() {
		adminPrivate.Path("api/v1/entities/{entityID}").HandlerFunc(handler.updateEntity()).Methods("POST")

		adminPrivate.Path("/entities/{id}").HandlerFunc(handler.updateEntityOld()).Methods("POST")
		adminPrivate.Path("/api/entities/{id}").HandlerFunc(handler.deleteEntity()).Methods("DELETE")
	})
}

func (handler *adminEntityHandler) updateEntityMemberStartedAt(oldEntity *types.Entity, newStatus string) {
	// Set timestamp when first trading status applied.
	if oldEntity.MemberStartedAt.IsZero() && (oldEntity.Status == constant.Entity.Accepted) && (newStatus == constant.Trading.Accepted) {
		logic.Entity.SetMemberStartedAt(oldEntity.ID)
	}
}

func (handler *adminEntityHandler) updateOfferAndWants(oldEntity *types.Entity, newStatus string, offers []string, wants []string) {
	tagDifference := types.NewTagDifference(types.TagFieldToNames(oldEntity.Offers), offers, types.TagFieldToNames(oldEntity.Wants), wants)
	err := logic.Entity.UpdateTags(oldEntity.ID, tagDifference)
	if err != nil {
		l.Logger.Error("[Error] AdminEntityHandler.updateOfferAndWants failed:", zap.Error(err))
		return
	}
	// Admin Update tags logic:
	// 	1. When a entity' status is changed from pending/rejected to accepted.
	// 	   - update all tags.
	// 	2. When the entity is in accepted status.
	//	   - only update added tags.
	if !util.IsAcceptedStatus(oldEntity.Status) && util.IsAcceptedStatus(newStatus) {
		err := logic.Entity.UpdateAllTagsCreatedAt(oldEntity.ID, time.Now())
		if err != nil {
			l.Logger.Error("[Error] AdminEntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
		err = TagHandler.SaveOfferTags(offers)
		if err != nil {
			l.Logger.Error("[Error] AdminEntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
		err = TagHandler.SaveWantTags(wants)
		if err != nil {
			l.Logger.Error("[Error] AdminEntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
	}
	if util.IsAcceptedStatus(oldEntity.Status) && util.IsAcceptedStatus(newStatus) {
		err := TagHandler.SaveOfferTags(tagDifference.OffersAdded)
		if err != nil {
			l.Logger.Error("[Error] AdminEntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
		err = TagHandler.SaveWantTags(tagDifference.WantsAdded)
		if err != nil {
			l.Logger.Error("[Error] AdminEntityHandler.updateOfferAndWants failed:", zap.Error(err))
		}
	}
}

func (handler *adminEntityHandler) updateEntity() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.EntityRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := types.NewAdminUpdateEntityReqBody(r)
		if err != nil {
			l.Logger.Info("[INFO] AdminEntityHandler.updateEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		errs := req.Validate()
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		oldEntity, err := logic.Entity.FindByID(req.EntityID)
		if err != nil {
			l.Logger.Error("[Error] AdminEntityHandler.updateEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		newEntity, err := logic.Entity.AdminFindOneAndUpdate(&types.Entity{
			ID:                 req.EntityID,
			EntityName:         req.EntityName,
			Email:              req.Email,
			EntityPhone:        req.EntityPhone,
			IncType:            req.IncType,
			CompanyNumber:      req.CompanyNumber,
			Website:            req.Website,
			Turnover:           req.Turnover,
			Description:        req.Description,
			LocationAddress:    req.LocationAddress,
			LocationCity:       req.LocationCity,
			LocationRegion:     req.LocationRegion,
			LocationPostalCode: req.LocationPostalCode,
			LocationCountry:    req.LocationCountry,
			Status:             req.Status,
		})
		if err != nil {
			l.Logger.Error("[Error] AdminEntityHandler.updateEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		// Update offers and wants are seperate processes.
		go handler.updateOfferAndWants(oldEntity, req.Status, req.Offers, req.Wants)
		go handler.updateEntityMemberStartedAt(oldEntity, req.Status)
		go CategoryHandler.Update(req.Categories)

		if len(req.Offers) != 0 {
			newEntity.Offers = types.ToTagFields(req.Offers)
		}
		if len(req.Wants) != 0 {
			newEntity.Wants = types.ToTagFields(req.Wants)
		}
		api.Respond(w, r, http.StatusOK, respond{Data: types.NewEntityRespondWithEmail(newEntity)})
	}
}

// TO BE REMOVED

func (a *adminEntityHandler) updateEntityOld() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("admin/entity")
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		d := helper.GetUpdateData(r)

		vars := mux.Vars(r)
		id := vars["id"]
		bID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			l.Logger.Error("UpdateEntity failed", zap.Error(err))
			t.Error(w, r, d, err)
			return
		}
		d.Entity.ID = bID

		errorMessages := validate.UpdateEntity(d.Entity)
		maxPosBal, err := strconv.ParseFloat(r.FormValue("max_pos_bal"), 64)
		if err != nil {
			errorMessages = append(errorMessages, "Max pos balance should be a number")
		}
		d.Balance.MaxPosBal = math.Abs(maxPosBal)
		maxNegBal, err := strconv.ParseFloat(r.FormValue("max_neg_bal"), 64)
		if err != nil {
			errorMessages = append(errorMessages, "Max neg balance should be a number")
		}
		if math.Abs(maxNegBal) == 0 {
			d.Balance.MaxNegBal = 0
		} else {
			d.Balance.MaxNegBal = math.Abs(maxNegBal)
		}

		// Check if the current balance has exceeded the input balances.
		account, err := logic.Account.FindByEntityID(bID.Hex())
		if err != nil {
			l.Logger.Error("UpdateEntity failed", zap.Error(err))
			t.Error(w, r, d, err)
			return
		}
		if account.Balance > d.Balance.MaxPosBal {
			errorMessages = append(errorMessages, "The current account balance ("+fmt.Sprintf("%.2f", account.Balance)+") has exceed your max pos balance input")
		}
		if account.Balance < -math.Abs(d.Balance.MaxNegBal) {
			errorMessages = append(errorMessages, "The current account balance ("+fmt.Sprintf("%.2f", account.Balance)+") has exceed your max neg balance input")
		}
		if len(errorMessages) > 0 {
			l.Logger.Error("UpdateEntity failed", zap.Error(err))
			t.Render(w, r, d, errorMessages)
			return
		}

		// Update Entity
		oldEntity, err := logic.Entity.FindByID(bID)
		if err != nil {
			l.Logger.Error("UpdateEntity failed", zap.Error(err))
			t.Error(w, r, d, err)
			return
		}
		offersAdded, offersRemoved := helper.TagDifference(d.Entity.Offers, oldEntity.Offers)
		d.Entity.OffersAdded = offersAdded
		d.Entity.OffersRemoved = offersRemoved
		wantsAdded, wantsRemoved := helper.TagDifference(d.Entity.Wants, oldEntity.Wants)
		d.Entity.WantsAdded = wantsAdded
		d.Entity.WantsRemoved = wantsRemoved
		err = logic.Entity.UpdateEntity(bID, d.Entity, true)
		if err != nil {
			l.Logger.Error("UpdateEntity failed", zap.Error(err))
			t.Error(w, r, d, err)
			return
		}

		// Update BalanceLimit
		oldBalance, err := logic.BalanceLimit.FindByAccountID(account.ID)
		if err != nil {
			l.Logger.Error("UpdateEntity failed", zap.Error(err))
			t.Error(w, r, d, err)
			return
		}
		err = logic.BalanceLimit.Update(account.ID, d.Balance.MaxPosBal, d.Balance.MaxNegBal)
		if err != nil {
			l.Logger.Error("UpdateEntity failed", zap.Error(err))
			t.Error(w, r, d, err)
			return
		}

		// Update the admin tags collection.
		go func() {
			err := CategoryHandler.SaveAdminTags(d.Entity.Categories)
			if err != nil {
				l.Logger.Error("saveAdminTags failed", zap.Error(err))
			}
		}()
		go func() {
			objID, _ := primitive.ObjectIDFromHex(r.Header.Get("userID"))
			adminUser, err := logic.AdminUser.FindByID(objID)
			user, err := logic.User.FindByEntityID(bID)
			if err != nil {
				l.Logger.Error("log.Admin.ModifyEntity failed", zap.Error(err))
				return
			}
			err = logic.UserAction.Log(log.Admin.ModifyEntity(adminUser, user, oldEntity, d.Entity, oldBalance, d.Balance))
			if err != nil {
				l.Logger.Error("log.Admin.ModifyEntity failed", zap.Error(err))
			}
		}()

		// Admin Update tags logic:
		// 	1. When a entity' status is changed from pending/rejected to accepted.
		// 	   - update all tags.
		// 	2. When the entity is in accepted status.
		//	   - only update added tags.
		go func() {
			if !util.IsAcceptedStatus(oldEntity.Status) && util.IsAcceptedStatus(d.Entity.Status) {
				err := logic.Entity.UpdateAllTagsCreatedAt(oldEntity.ID, time.Now())
				if err != nil {
					l.Logger.Error("UpdateAllTagsCreatedAt failed", zap.Error(err))
				}
				err = TagHandler.SaveOfferTags(helper.GetTagNames(d.Entity.Offers))
				if err != nil {
					l.Logger.Error("saveOfferTags failed", zap.Error(err))
				}
				err = TagHandler.SaveWantTags(helper.GetTagNames(d.Entity.Wants))
				if err != nil {
					l.Logger.Error("saveWantTags failed", zap.Error(err))
				}
			}
			if util.IsAcceptedStatus(oldEntity.Status) && util.IsAcceptedStatus(d.Entity.Status) {
				err := TagHandler.SaveOfferTags(d.Entity.OffersAdded)
				if err != nil {
					l.Logger.Error("saveOfferTags failed", zap.Error(err))
				}
				err = TagHandler.SaveWantTags(d.Entity.WantsAdded)
				if err != nil {
					l.Logger.Error("saveWantTags failed", zap.Error(err))
				}
			}
		}()
		go func() {
			// Set timestamp when first trading status applied.
			if oldEntity.MemberStartedAt.IsZero() && (oldEntity.Status == constant.Entity.Accepted) && (d.Entity.Status == constant.Trading.Accepted) {
				logic.Entity.SetMemberStartedAt(bID)
			}
		}()

		t.Success(w, r, d, "The entity has been updated!")
	}
}

func (a *adminEntityHandler) deleteEntity() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		bsID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			l.Logger.Error("DeleteEntity failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = logic.Entity.DeleteByID(bsID)
		if err != nil {
			l.Logger.Error("DeleteEntity failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user, err := logic.User.FindByEntityID(bsID)
		if err != nil {
			l.Logger.Error("DeleteEntity failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = logic.User.DeleteByID(user.ID)
		if err != nil {
			l.Logger.Error("DeleteEntity failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
