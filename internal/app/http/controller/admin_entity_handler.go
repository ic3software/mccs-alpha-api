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
	"github.com/ic3network/mccs-alpha-api/internal/pkg/helper"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/log"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/template"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/utils"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/validate"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
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

func (a *adminEntityHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	a.once.Do(func() {
		adminPrivate.Path("/entities/{id}").HandlerFunc(a.adminEntityPage()).Methods("GET")
		adminPrivate.Path("/entities/{id}").HandlerFunc(a.updateEntity()).Methods("POST")

		adminPrivate.Path("/api/entities/{id}").HandlerFunc(a.deleteEntity()).Methods("DELETE")
	})
}

func (a *adminEntityHandler) adminEntityPage() func(http.ResponseWriter, *http.Request) {
	t := template.NewView("admin/entity")
	type formData struct {
		Entity  *types.Entity
		Balance *types.BalanceLimit
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		entity, err := EntityHandler.FindByID(id)
		if err != nil {
			l.Logger.Error("AdminEntityPage failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}
		balance, err := logic.BalanceLimit.FindByEntityID(id)
		if err != nil {
			l.Logger.Error("AdminEntityPage failed", zap.Error(err))
			t.Error(w, r, nil, err)
			return
		}

		f := formData{Entity: entity, Balance: balance}

		t.Render(w, r, f, nil)
	}
}

func (a *adminEntityHandler) updateEntity() func(http.ResponseWriter, *http.Request) {
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
			if !utils.IsAcceptedStatus(oldEntity.Status) && utils.IsAcceptedStatus(d.Entity.Status) {
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
			if utils.IsAcceptedStatus(oldEntity.Status) && utils.IsAcceptedStatus(d.Entity.Status) {
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
