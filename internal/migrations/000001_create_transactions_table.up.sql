CREATE TABLE public.transactions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    order_id character varying(255) NOT NULL,
    transaction_date timestamp(0) without time zone NOT NULL,
    amount numeric(15,2) NOT NULL,
    payment_method character varying(255) NOT NULL,
    status character varying(255) NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone,

    CONSTRAINT transactions_payment_method_check CHECK (((payment_method)::text = ANY ((ARRAY['credit_card'::character varying, 'debit_card'::character varying, 'e_wallet'::character varying, 'bank_transfer'::character varying, 'cash_on_delivery'::character varying])::text[]))),

    CONSTRAINT transactions_status_check CHECK (((status)::text = ANY ((ARRAY['completed'::character varying, 'pending'::character varying, 'canceled'::character varying, 'failed'::character varying])::text[])))
);
