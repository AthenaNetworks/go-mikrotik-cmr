
void FUN_08059380(void)

{
  BigNum *in_EAX;
  Point *in_EDX;
  ECPoint *pEVar1;
  ECPoint local_40 [48];
  
  pEVar1 = local_40;
  WCurve::fromEC(pEVar1);
  WCurve::mul(in_EAX,in_EDX);
  FUN_08059352(pEVar1);
  return;
}


