// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package db

import (
	"context"
)

type Querier interface {
	CountActions(ctx context.Context) (int64, error)
	CountActionsByPersonID(ctx context.Context, xidStr string) (int64, error)
	CountPersons(ctx context.Context) (int64, error)
	CreateAction(ctx context.Context, arg CreateActionParams) (CreateActionRow, error)
	CreatePerson(ctx context.Context, arg CreatePersonParams) (CreatePersonRow, error)
	DeleteAction(ctx context.Context, xidStr string) error
	DeletePerson(ctx context.Context, xidStr string) error
	GetActionByID(ctx context.Context, xidStr string) (GetActionByIDRow, error)
	GetActionsByDateRange(ctx context.Context, arg GetActionsByDateRangeParams) ([]GetActionsByDateRangeRow, error)
	GetActionsWithPersonDetails(ctx context.Context, arg GetActionsWithPersonDetailsParams) ([]GetActionsWithPersonDetailsRow, error)
	GetPersonByID(ctx context.Context, xidStr string) (GetPersonByIDRow, error)
	GetPersonByName(ctx context.Context, name string) (GetPersonByNameRow, error)
	GetRecentActionsByPersonID(ctx context.Context, arg GetRecentActionsByPersonIDParams) ([]GetRecentActionsByPersonIDRow, error)
	ListActions(ctx context.Context, arg ListActionsParams) ([]ListActionsRow, error)
	ListActionsByPersonID(ctx context.Context, arg ListActionsByPersonIDParams) ([]ListActionsByPersonIDRow, error)
	ListActionsByPersonIDAndValence(ctx context.Context, arg ListActionsByPersonIDAndValenceParams) ([]ListActionsByPersonIDAndValenceRow, error)
	ListActionsByValence(ctx context.Context, arg ListActionsByValenceParams) ([]ListActionsByValenceRow, error)
	ListPersons(ctx context.Context, arg ListPersonsParams) ([]ListPersonsRow, error)
	ListPersonsWithLastAction(ctx context.Context, arg ListPersonsWithLastActionParams) ([]ListPersonsWithLastActionRow, error)
	SearchActionsByDescription(ctx context.Context, arg SearchActionsByDescriptionParams) ([]SearchActionsByDescriptionRow, error)
	SearchPersonsByName(ctx context.Context, arg SearchPersonsByNameParams) ([]SearchPersonsByNameRow, error)
	UpdateAction(ctx context.Context, arg UpdateActionParams) (UpdateActionRow, error)
	UpdatePerson(ctx context.Context, arg UpdatePersonParams) (UpdatePersonRow, error)
}

var _ Querier = (*Queries)(nil)
