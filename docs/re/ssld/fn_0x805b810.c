
char FUN_0805b810(void)

{
  int in_EAX;
  undefined4 uVar1;
  Rng *in_ECX;
  BigNum *pBVar2;
  char local_6d;
  undefined1 local_69;
  BigNum local_68 [16];
  BigNum local_58 [24];
  undefined1 *local_40;
  undefined4 local_3c;
  
  local_6d = FUN_080593c0();
  if (local_6d != '\0') {
    local_40 = &local_69;
    local_69 = 1;
    local_3c = 1;
    FUN_0804f6aa();
    uVar1 = FUN_08050e36();
    BigNum::operator-=(local_68,local_58);
    pBVar2 = local_68;
    local_6d = randomBigNum((BigNum *)(in_EAX + 0x3c),local_68,in_ECX);
    FUN_0804edbe(local_68,pBVar2,in_ECX,uVar1);
    FUN_0804edbe(local_58);
    if (local_6d != '\0') {
      FUN_0805b76a();
      FUN_08059380(in_EAX + 0x18);
      WCurve::toEC((Point *)local_68);
      FUN_0804fc9a();
      FUN_0804fc9a();
      FUN_0804f6cc();
      FUN_08059352();
      FUN_0805931c();
    }
  }
  return local_6d;
}


